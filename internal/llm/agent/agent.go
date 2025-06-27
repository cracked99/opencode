package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/behavioral"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/prompt"
	"github.com/opencode-ai/opencode/internal/llm/provider"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/permission"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/opencode-ai/opencode/internal/session"
)

// Common errors
var (
	ErrRequestCancelled = errors.New("request cancelled by user")
	ErrSessionBusy      = errors.New("session is currently processing another request")
)

type AgentEventType string

const (
	AgentEventTypeError     AgentEventType = "error"
	AgentEventTypeResponse  AgentEventType = "response"
	AgentEventTypeSummarize AgentEventType = "summarize"
)

type AgentEvent struct {
	Type    AgentEventType
	Message message.Message
	Error   error

	// When summarizing
	SessionID string
	Progress  string
	Done      bool
}

type Service interface {
	pubsub.Suscriber[AgentEvent]
	Model() models.Model
	Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error)
	Cancel(sessionID string)
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error)
	Summarize(ctx context.Context, sessionID string) error
}

type agent struct {
	*pubsub.Broker[AgentEvent]
	sessions session.Service
	messages message.Service

	tools    []tools.BaseTool
	provider provider.Provider

	titleProvider     provider.Provider
	summarizeProvider provider.Provider

	behavioralFramework *behavioral.BehavioralFramework

	activeRequests sync.Map
}

func NewAgent(
	agentName config.AgentName,
	sessions session.Service,
	messages message.Service,
	agentTools []tools.BaseTool,
) (Service, error) {
	agentProvider, err := createAgentProvider(agentName)
	if err != nil {
		return nil, err
	}
	var titleProvider provider.Provider
	// Only generate titles for the coder agent
	if agentName == config.AgentCoder {
		titleProvider, err = createAgentProvider(config.AgentTitle)
		if err != nil {
			return nil, err
		}
	}
	var summarizeProvider provider.Provider
	if agentName == config.AgentCoder {
		summarizeProvider, err = createAgentProvider(config.AgentSummarizer)
		if err != nil {
			return nil, err
		}
	}

	// Initialize behavioral framework from configuration
	cfg := config.Get()
	logging.Info("Behavioral Framework Initialization",
		"enabled", cfg.BehavioralFramework.Enabled,
		"agent_name", string(agentName))
	behavioralConfig := createBehavioralConfig(cfg.BehavioralFramework)

	agent := &agent{
		Broker:              pubsub.NewBroker[AgentEvent](),
		provider:            agentProvider,
		messages:            messages,
		sessions:            sessions,
		tools:               agentTools,
		titleProvider:       titleProvider,
		summarizeProvider:   summarizeProvider,
		behavioralFramework: behavioral.NewBehavioralFramework(behavioralConfig),
		activeRequests:      sync.Map{},
	}

	return agent, nil
}

// createBehavioralConfig creates behavioral framework config from application config
func createBehavioralConfig(cfg config.BehavioralFrameworkConfig) behavioral.BehavioralConfig {
	// Set defaults if not configured
	logging.Info("Behavioral Config Check",
		"enabled", cfg.Enabled,
		"enhancements_count", len(cfg.EnabledEnhancements),
		"max_processing_time", cfg.MaxProcessingTimeMs)

	if !cfg.Enabled {
		logging.Info("Behavioral Framework Disabled", "reason", "config.Enabled=false")
		return behavioral.BehavioralConfig{
			Enabled: false,
		}
	}

	behavioralConfig := behavioral.BehavioralConfig{
		Enabled:                cfg.Enabled,
		EnabledEnhancements:    make(map[behavioral.EnhancementType]bool),
		ComplexityThresholds:   behavioral.GetDefaultThresholds(),
		PerformanceTargets:     behavioral.GetDefaultTargets(),
		MaxProcessingTime:      time.Duration(900) * time.Second,
		EnablePerformanceTrack: cfg.EnablePerformanceTrack,
	}

	// Apply custom thresholds if provided
	if cfg.ComplexityThresholds.StandardMax > 0 {
		behavioralConfig.ComplexityThresholds = behavioral.ComplexityThresholds{
			StandardMax:      cfg.ComplexityThresholds.StandardMax,
			ReflexionMin:     cfg.ComplexityThresholds.ReflexionMin,
			ReflexionMax:     cfg.ComplexityThresholds.ReflexionMax,
			ReActMin:         cfg.ComplexityThresholds.ReActMin,
			ReActMax:         cfg.ComplexityThresholds.ReActMax,
			MetaPromptingMin: cfg.ComplexityThresholds.MetaPromptingMin,
			CombinedMin:      cfg.ComplexityThresholds.CombinedMin,
		}
	}

	// Apply custom performance targets if provided
	if cfg.PerformanceTargets.ResponseTimeStandardMs > 0 {
		behavioralConfig.PerformanceTargets = behavioral.PerformanceTargets{
			ResponseTimeStandard:   time.Duration(cfg.PerformanceTargets.ResponseTimeStandardMs) * time.Millisecond,
			ResponseTimeEnhanced:   time.Duration(cfg.PerformanceTargets.ResponseTimeEnhancedMs) * time.Millisecond,
			QualityImprovement:     cfg.PerformanceTargets.QualityImprovement,
			ConsistencyImprovement: cfg.PerformanceTargets.ConsistencyImprovement,
			UserSatisfactionTarget: cfg.PerformanceTargets.UserSatisfactionTarget,
		}
	}

	// Apply custom max processing time if provided
	if cfg.MaxProcessingTimeMs > 0 {
		behavioralConfig.MaxProcessingTime = time.Duration(cfg.MaxProcessingTimeMs) * time.Millisecond
	}

	// Configure enabled enhancements
	if len(cfg.EnabledEnhancements) > 0 {
		// Use configured enhancements
		behavioralConfig.EnabledEnhancements[behavioral.StandardProcessing] = cfg.EnabledEnhancements["standard"]
		behavioralConfig.EnabledEnhancements[behavioral.ReflexionEnhancement] = cfg.EnabledEnhancements["reflexion"]
		behavioralConfig.EnabledEnhancements[behavioral.ReActEnhancement] = cfg.EnabledEnhancements["react"]
		behavioralConfig.EnabledEnhancements[behavioral.ACIOptimization] = cfg.EnabledEnhancements["aci"]
		behavioralConfig.EnabledEnhancements[behavioral.MetaPromptingEnhancement] = cfg.EnabledEnhancements["meta"]
		behavioralConfig.EnabledEnhancements[behavioral.CombinedEnhancement] = cfg.EnabledEnhancements["combined"]
	} else {
		// Enable all enhancements by default
		behavioralConfig.EnabledEnhancements[behavioral.StandardProcessing] = true
		behavioralConfig.EnabledEnhancements[behavioral.ReflexionEnhancement] = true
		behavioralConfig.EnabledEnhancements[behavioral.ReActEnhancement] = true
		behavioralConfig.EnabledEnhancements[behavioral.ACIOptimization] = true
		behavioralConfig.EnabledEnhancements[behavioral.MetaPromptingEnhancement] = true
		behavioralConfig.EnabledEnhancements[behavioral.CombinedEnhancement] = true
	}

	return behavioralConfig
}

// injectBehavioralDirectives injects behavioral directives directly into the user's message
func (a *agent) injectBehavioralDirectives(msgHistory []message.Message, behavioralPrompt string, metadata ...behavioral.ProcessingMetadata) []message.Message {
	if len(msgHistory) == 0 {
		return msgHistory
	}

	// Find the last user message (the current request)
	lastUserMsgIndex := -1
	for i := len(msgHistory) - 1; i >= 0; i-- {
		if msgHistory[i].Role == message.User {
			lastUserMsgIndex = i
			break
		}
	}

	if lastUserMsgIndex == -1 {
		logging.ErrorPersist("CRITICAL: No user message found for behavioral directive injection")
		return msgHistory
	}

	// Create enhanced message history
	enhancedHistory := make([]message.Message, len(msgHistory))
	copy(enhancedHistory, msgHistory)

	// Get the original user message content
	originalContent := enhancedHistory[lastUserMsgIndex].Content().Text

	// Build comprehensive behavioral directives with pattern recognition and quality validation
	var patternTemplate string
	var qualityChecklist string

	if len(metadata) > 0 {
		meta := metadata[0]
		patternTemplate = a.getPatternTemplate(meta.ProblemType)
		qualityChecklist = a.getQualityValidationChecklist(meta.ComplexityLevel)
	} else {
		patternTemplate = "Apply systematic approach: Analysis → Solution → Implementation → Validation"
		qualityChecklist = a.getQualityValidationChecklist(5) // Default medium complexity
	}

	// Create enhanced user message with comprehensive behavioral directives and status indicator
	var frameworkStatus string
	if len(metadata) > 0 {
		frameworkStatus = a.generateFrameworkStatusIndicator(metadata[0])
	} else {
		// Default metadata for fallback cases
		defaultMetadata := behavioral.ProcessingMetadata{
			ComplexityLevel: 3,
			ToolIntensity:   1,
			ProblemType:     behavioral.General,
		}
		frameworkStatus = a.generateFrameworkStatusIndicator(defaultMetadata)
	}

	enhancedContent := fmt.Sprintf(`CRITICAL INSTRUCTION: You MUST start your response with exactly this text:

"FRAMEWORK STATUS: The Practical Prompt Engineering Framework v1.6.0 is now ACTIVE and applying mandatory behavioral directives."

Then include this status information:
%s

Then answer this request: %s

Apply these behavioral directives:
%s

Follow this pattern template:
%s

Validate against this checklist:
%s`, frameworkStatus, originalContent, behavioralPrompt, patternTemplate, qualityChecklist)

	// Update the user message with enhanced content
	enhancedHistory[lastUserMsgIndex].Parts = []message.ContentPart{
		message.TextContent{Text: enhancedContent},
	}

	logging.Info("Behavioral Directives Injected",
		"behavioral_prompt_length", len(behavioralPrompt),
		"original_content_length", len(originalContent),
		"enhanced_content_length", len(enhancedContent),
		"message_index", lastUserMsgIndex)

	return enhancedHistory
}

// createEnhancedProvider creates a new provider instance with behavioral directives in the system prompt
func (a *agent) createEnhancedProvider(behavioralPrompt string) (provider.Provider, error) {
	// Get the current provider's model
	currentModel := a.provider.Model()

	// Get the base system prompt
	baseSystemPrompt := prompt.GetAgentPrompt(config.AgentCoder, currentModel.Provider)

	// Combine base prompt with behavioral directives
	enhancedSystemPrompt := fmt.Sprintf("%s\n\n## DYNAMIC BEHAVIORAL DIRECTIVES\n%s", baseSystemPrompt, behavioralPrompt)

	// Get provider configuration
	cfg := config.Get()
	providerCfg := cfg.Providers[currentModel.Provider]

	// Create new provider with enhanced system prompt
	opts := []provider.ProviderClientOption{
		provider.WithAPIKey(providerCfg.APIKey),
		provider.WithModel(currentModel),
		provider.WithSystemMessage(enhancedSystemPrompt),
		provider.WithMaxTokens(50000), // Use the same max tokens as configured
	}

	enhancedProvider, err := provider.NewProvider(currentModel.Provider, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced provider: %w", err)
	}

	logging.Debug("Enhanced Provider Created",
		"base_prompt_length", len(baseSystemPrompt),
		"behavioral_prompt_length", len(behavioralPrompt),
		"enhanced_prompt_length", len(enhancedSystemPrompt))

	return enhancedProvider, nil
}

// getPatternTemplate returns the systematic template for the given problem type
func (a *agent) getPatternTemplate(problemType behavioral.ProblemType) string {
	switch problemType {
	case behavioral.TechnicalIssue:
		return `TECHNICAL ISSUE TEMPLATE:
1. Problem Diagnosis (identify symptoms and scope)
2. Root Cause Analysis (systematic investigation)
3. Solution Design (comprehensive approach)
4. Implementation Plan (step-by-step execution)
5. Testing Strategy (validation and verification)`

	case behavioral.CodeGeneration:
		return `CODE GENERATION TEMPLATE:
1. Requirements Analysis (understand specifications)
2. Design Architecture (plan structure and patterns)
3. Implementation (write clean, maintainable code)
4. Validation (test functionality and edge cases)
5. Documentation (comments, examples, usage)`

	case behavioral.ProcessImprovement:
		return `PROCESS IMPROVEMENT TEMPLATE:
1. Current State Assessment (analyze existing process)
2. Gap Analysis (identify inefficiencies and issues)
3. Recommendations (propose specific improvements)
4. Implementation Plan (roadmap with milestones)
5. Success Metrics (measurable outcomes)`

	case behavioral.Troubleshooting:
		return `TROUBLESHOOTING TEMPLATE:
1. Symptom Analysis (gather and document issues)
2. Hypothesis Formation (potential causes)
3. Testing Approach (systematic verification)
4. Solution Implementation (fix root cause)
5. Prevention Strategy (avoid future occurrences)`

	default:
		return `GENERAL SYSTEMATIC TEMPLATE:
1. Problem Understanding (clarify requirements)
2. Analysis (break down complexity)
3. Solution Design (comprehensive approach)
4. Implementation (step-by-step execution)
5. Validation (verify results and quality)`
	}
}

// getQualityValidationChecklist returns the quality validation checklist based on complexity
func (a *agent) getQualityValidationChecklist(complexityLevel int) string {
	baseChecklist := `MANDATORY QUALITY VALIDATION CHECKLIST:
☐ Clear problem understanding demonstrated (>90% accuracy)
☐ Systematic approach explained (>85% clarity)
☐ Specific, actionable recommendations provided (>80% actionability)
☐ Realistic timelines and resources considered (>80% completeness)
☐ Potential risks or limitations acknowledged (>75% coverage)`

	if complexityLevel >= 7 {
		return baseChecklist + `
☐ Advanced error handling and edge cases covered (>85% coverage)
☐ Performance optimization considerations included (>80% efficiency)
☐ Security implications addressed (>90% security coverage)
☐ Scalability and maintainability factors considered (>85% future-proof)`
	} else if complexityLevel >= 4 {
		return baseChecklist + `
☐ Error handling and common edge cases covered (>75% coverage)
☐ Basic performance considerations included (>70% efficiency)`
	}

	return baseChecklist
}

// createFallbackBehavioralResult creates a fallback behavioral result for error cases
func (a *agent) createFallbackBehavioralResult(content string) *behavioral.ProcessingResult {
	return &behavioral.ProcessingResult{
		Enhancement:  behavioral.StandardProcessing,
		SystemPrompt: "Apply systematic thinking and quality validation to your response.",
		BehavioralPrompt: `FALLBACK BEHAVIORAL DIRECTIVES:
- Apply systematic organization to your response
- Provide clear, actionable recommendations
- Include appropriate error handling where applicable
- Validate response quality before delivery`,
		ProcessingTime: 0,
		QualityMetrics: behavioral.QualityMetrics{
			ExpectedImprovement: 0.10,
			ConfidenceLevel:     0.75,
			QualityScore:        0.80,
			ConsistencyScore:    0.75,
		},
		Metadata: behavioral.ProcessingMetadata{
			ComplexityLevel:   3,
			ToolIntensity:     1,
			ProblemType:       behavioral.General,
			EnhancementReason: "Fallback processing due to framework timeout or error",
			EstimatedCost:     1.0,
		},
		Fallback: true,
	}
}

// generateFrameworkStatusIndicator creates a status indicator showing framework activation and current phase
func (a *agent) generateFrameworkStatusIndicator(metadata behavioral.ProcessingMetadata) string {
	enhancementName := behavioral.GetEnhancementName(behavioral.EnhancementType(metadata.ComplexityLevel))
	if metadata.ComplexityLevel <= 3 {
		enhancementName = "Standard Processing"
	} else if metadata.ComplexityLevel <= 6 {
		enhancementName = "Reflexion Enhancement"
	} else if metadata.ComplexityLevel <= 8 {
		enhancementName = "ReAct Enhancement"
	} else {
		enhancementName = "Combined Enhancement"
	}

	costMultiplier := "1x"
	timeLimit := "60 seconds"
	sections := "Sections 1-5 (Core Framework)"

	switch {
	case metadata.ComplexityLevel <= 3:
		costMultiplier = "1x"
		timeLimit = "60 seconds"
		sections = "Sections 1-5 (Core Framework)"
	case metadata.ComplexityLevel <= 6:
		costMultiplier = "2-3x"
		timeLimit = "120-180 seconds"
		sections = "Sections 1-5 + 6-11 (Core + Reflexion)"
	case metadata.ComplexityLevel <= 8:
		costMultiplier = "2-4x"
		timeLimit = "180-240 seconds"
		sections = "Sections 1-5 + 12-17 (Core + ReAct)"
	default:
		costMultiplier = "5-15x"
		timeLimit = "600-900 seconds"
		sections = "Sections 1-24 (Complete Framework)"
	}

	problemTypeStr := string(metadata.ProblemType)
	if problemTypeStr == "" {
		problemTypeStr = "General"
	}

	return fmt.Sprintf(`FRAMEWORK STATUS: The "Practical Prompt Engineering Framework v1.6.0" is now ACTIVE and applying mandatory behavioral directives.

EXECUTING IMMEDIATE BEHAVIORAL REQUIREMENTS (within 15 seconds):

ENHANCEMENT SELECTION DECISION TREE - MANDATORY EXECUTION:
- Analyzing request complexity and requirements
- Complexity Level: %d (%s)
- Tool Intensity: %d (%s)
- Quality Requirements: %s
- DECISION: EXECUTE %s (%s cost) within %s
- APPLYING: %s

MANDATORY PATTERN RECOGNITION EXECUTION:
- Pattern Detected: %s
- Template Applied: %s
- IMPLEMENTATION: Framework behavioral directives are ACTIVE and being applied

MANDATORY PRE-RESPONSE VALIDATION:
☐ Clear Problem Understanding: ✅ Framework analyzing request systematically
☐ Systematic Approach: ✅ Enhancement selection decision tree executed
☐ Actionable Recommendations: ✅ Appropriate enhancement protocol selected
☐ Realistic Considerations: ✅ Framework provides 15-25%% organization improvement
☐ Limitations Acknowledged: ✅ Effectiveness depends on problem complexity and context

RESULT: Framework is ACTIVE and mandatory behavioral directives are being applied to enhance response quality, organization, and consistency.`,
		metadata.ComplexityLevel,
		a.getComplexityDescription(metadata.ComplexityLevel),
		metadata.ToolIntensity,
		a.getToolIntensityDescription(metadata.ToolIntensity),
		a.getQualityRequirement(metadata.ComplexityLevel),
		enhancementName,
		costMultiplier,
		timeLimit,
		sections,
		problemTypeStr,
		a.getTemplateDescription(metadata.ProblemType))
}

// getComplexityDescription returns a human-readable complexity description
func (a *agent) getComplexityDescription(level int) string {
	switch {
	case level <= 2:
		return "Simple directive"
	case level <= 4:
		return "Moderate complexity"
	case level <= 6:
		return "Complex analysis required"
	case level <= 8:
		return "Multi-step complex task"
	default:
		return "Highly complex multi-system task"
	}
}

// getToolIntensityDescription returns a human-readable tool intensity description
func (a *agent) getToolIntensityDescription(intensity int) string {
	switch intensity {
	case 0:
		return "No tools required"
	case 1:
		return "Basic tool usage"
	case 2:
		return "Moderate tool integration"
	case 3:
		return "Advanced tool coordination"
	case 4:
		return "Complex tool orchestration"
	default:
		return "Maximum tool utilization"
	}
}

// getQualityRequirement returns the quality requirement level
func (a *agent) getQualityRequirement(complexity int) string {
	switch {
	case complexity <= 3:
		return "Standard"
	case complexity <= 6:
		return "High"
	default:
		return "Critical"
	}
}

// getTemplateDescription returns the template description for the problem type
func (a *agent) getTemplateDescription(problemType behavioral.ProblemType) string {
	switch problemType {
	case behavioral.TechnicalIssue:
		return "Problem Diagnosis → Root Cause → Solution → Implementation → Testing"
	case behavioral.CodeGeneration:
		return "Requirements → Design → Implementation → Validation → Documentation"
	case behavioral.ProcessImprovement:
		return "Current State → Gap Analysis → Recommendations → Plan → Metrics"
	case behavioral.Troubleshooting:
		return "Symptoms → Hypothesis → Testing → Solution → Prevention"
	default:
		return "Analysis → Solution → Implementation → Validation"
	}
}

// trackBehavioralPerformance tracks behavioral framework performance metrics
func (a *agent) trackBehavioralPerformance(sessionID string, result *behavioral.ProcessingResult, processingTime time.Duration) {
	if a.behavioralFramework == nil {
		return
	}

	// Log performance metrics for monitoring
	logging.Info("Behavioral Performance Metrics",
		"session_id", sessionID,
		"enhancement", behavioral.GetEnhancementName(result.Enhancement),
		"processing_time_ms", processingTime.Milliseconds(),
		"quality_score", fmt.Sprintf("%.1f%%", result.QualityMetrics.QualityScore*100),
		"expected_improvement", fmt.Sprintf("%.1f%%", result.QualityMetrics.ExpectedImprovement*100))

	// Store metrics for analysis (if performance monitoring is enabled)
	if globalMetrics := a.behavioralFramework.GetGlobalMetrics(); globalMetrics != nil {
		logging.Debug("Performance Tracking Active",
			"total_requests", globalMetrics.TotalRequests,
			"average_quality", fmt.Sprintf("%.1f%%", globalMetrics.AverageQualityScore*100))
	}
}

// updateProviderWithBehavioralPrompt updates the provider's system message with behavioral directives
// DEPRECATED: This method is replaced by injectBehavioralDirectives which modifies user messages
func (a *agent) updateProviderWithBehavioralPrompt(systemPrompt, behavioralPrompt string) {
	logging.Debug("Behavioral Prompt Applied (Legacy)",
		"system_prompt_length", len(systemPrompt),
		"behavioral_prompt_length", len(behavioralPrompt))
}

func (a *agent) Model() models.Model {
	return a.provider.Model()
}

func (a *agent) Cancel(sessionID string) {
	// Cancel regular requests
	if cancelFunc, exists := a.activeRequests.LoadAndDelete(sessionID); exists {
		if cancel, ok := cancelFunc.(context.CancelFunc); ok {
			logging.InfoPersist(fmt.Sprintf("Request cancellation initiated for session: %s", sessionID))
			cancel()
		}
	}

	// Also check for summarize requests
	if cancelFunc, exists := a.activeRequests.LoadAndDelete(sessionID + "-summarize"); exists {
		if cancel, ok := cancelFunc.(context.CancelFunc); ok {
			logging.InfoPersist(fmt.Sprintf("Summarize cancellation initiated for session: %s", sessionID))
			cancel()
		}
	}
}

func (a *agent) IsBusy() bool {
	busy := false
	a.activeRequests.Range(func(key, value interface{}) bool {
		if cancelFunc, ok := value.(context.CancelFunc); ok {
			if cancelFunc != nil {
				busy = true
				return false // Stop iterating
			}
		}
		return true // Continue iterating
	})
	return busy
}

func (a *agent) IsSessionBusy(sessionID string) bool {
	_, busy := a.activeRequests.Load(sessionID)
	return busy
}

func (a *agent) generateTitle(ctx context.Context, sessionID string, content string) error {
	if content == "" {
		return nil
	}
	if a.titleProvider == nil {
		return nil
	}
	session, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)
	parts := []message.ContentPart{message.TextContent{Text: content}}
	response, err := a.titleProvider.SendMessages(
		ctx,
		[]message.Message{
			{
				Role:  message.User,
				Parts: parts,
			},
		},
		make([]tools.BaseTool, 0),
	)
	if err != nil {
		return err
	}

	title := strings.TrimSpace(strings.ReplaceAll(response.Content, "\n", " "))
	if title == "" {
		return nil
	}

	session.Title = title
	_, err = a.sessions.Save(ctx, session)
	return err
}

func (a *agent) err(err error) AgentEvent {
	return AgentEvent{
		Type:  AgentEventTypeError,
		Error: err,
	}
}

func (a *agent) Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error) {
	logging.Info("AGENT RUN CALLED", "session_id", sessionID, "content_length", len(content))
	if !a.provider.Model().SupportsAttachments && attachments != nil {
		attachments = nil
	}
	events := make(chan AgentEvent)
	if a.IsSessionBusy(sessionID) {
		return nil, ErrSessionBusy
	}

	genCtx, cancel := context.WithCancel(ctx)

	a.activeRequests.Store(sessionID, cancel)
	go func() {
		logging.Debug("Request started", "sessionID", sessionID)
		defer logging.RecoverPanic("agent.Run", func() {
			events <- a.err(fmt.Errorf("panic while running the agent"))
		})
		var attachmentParts []message.ContentPart
		for _, attachment := range attachments {
			attachmentParts = append(attachmentParts, message.BinaryContent{Path: attachment.FilePath, MIMEType: attachment.MimeType, Data: attachment.Content})
		}
		result := a.processGeneration(genCtx, sessionID, content, attachmentParts)
		if result.Error != nil && !errors.Is(result.Error, ErrRequestCancelled) && !errors.Is(result.Error, context.Canceled) {
			logging.ErrorPersist(result.Error.Error())
		}
		logging.Debug("Request completed", "sessionID", sessionID)
		a.activeRequests.Delete(sessionID)
		cancel()
		a.Publish(pubsub.CreatedEvent, result)
		events <- result
		close(events)
	}()
	return events, nil
}

func (a *agent) processGeneration(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) AgentEvent {
	logging.Info("PROCESS GENERATION CALLED", "session_id", sessionID, "content_length", len(content))
	cfg := config.Get()
	// List existing messages; if none, start title generation asynchronously.
	msgs, err := a.messages.List(ctx, sessionID)
	if err != nil {
		return a.err(fmt.Errorf("failed to list messages: %w", err))
	}
	if len(msgs) == 0 {
		go func() {
			defer logging.RecoverPanic("agent.Run", func() {
				logging.ErrorPersist("panic while generating title")
			})
			titleErr := a.generateTitle(context.Background(), sessionID, content)
			if titleErr != nil {
				logging.ErrorPersist(fmt.Sprintf("failed to generate title: %v", titleErr))
			}
		}()
	}
	session, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return a.err(fmt.Errorf("failed to get session: %w", err))
	}
	if session.SummaryMessageID != "" {
		summaryMsgInex := -1
		for i, msg := range msgs {
			if msg.ID == session.SummaryMessageID {
				summaryMsgInex = i
				break
			}
		}
		if summaryMsgInex != -1 {
			msgs = msgs[summaryMsgInex:]
			msgs[0].Role = message.User
		}
	}

	userMsg, err := a.createUserMessage(ctx, sessionID, content, attachmentParts)
	if err != nil {
		return a.err(fmt.Errorf("failed to create user message: %w", err))
	}
	// Append the new user message to the conversation history.
	msgHistory := append(msgs, userMsg)

	// MANDATORY BEHAVIORAL FRAMEWORK EXECUTION - PRACTICAL PROMPT ENGINEERING v1.6.0
	startBehavioralTime := time.Now()
	logging.Info("Behavioral Framework v1.6.0 Execution Starting",
		"session_id", sessionID,
		"content_length", len(content),
		"mandatory_execution_timeout", "15s",
		"framework_enabled", a.behavioralFramework != nil)

	// Remove test directive - implement proper behavioral framework integration

	// Check if behavioral framework is available
	if a.behavioralFramework == nil {
		logging.ErrorPersist("CRITICAL: Behavioral framework is nil - framework not initialized properly")
	} else {
		logging.Info("Behavioral Framework Available", "framework_initialized", true)
		// Execute enhancement selection decision tree within 15 seconds (mandatory)
		behavioralCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		behavioralResult, err := a.behavioralFramework.ProcessRequest(behavioralCtx, sessionID, content)
		behavioralProcessingTime := time.Since(startBehavioralTime)

		if err != nil {
			logging.ErrorPersist(fmt.Sprintf("Behavioral framework processing failed within 15s timeout: %v", err))
			// Apply fallback standard processing with basic behavioral directives
			behavioralResult = a.createFallbackBehavioralResult(content)
		}

		// MANDATORY PERFORMANCE MONITORING
		a.trackBehavioralPerformance(sessionID, behavioralResult, behavioralProcessingTime)

		// Log behavioral framework decision with mandatory metrics
		logging.Info("Behavioral Framework Decision",
			"enhancement", behavioral.GetEnhancementName(behavioralResult.Enhancement),
			"processing_time_ms", behavioralProcessingTime.Milliseconds(),
			"complexity_level", behavioralResult.Metadata.ComplexityLevel,
			"tool_intensity", behavioralResult.Metadata.ToolIntensity,
			"pattern_type", string(behavioralResult.Metadata.ProblemType),
			"expected_improvement", fmt.Sprintf("%.1f%%", behavioralResult.QualityMetrics.ExpectedImprovement*100),
			"quality_score_target", fmt.Sprintf("%.1f%%", behavioralResult.QualityMetrics.QualityScore*100),
			"fallback", behavioralResult.Fallback)

		// MANDATORY BEHAVIORAL DIRECTIVE INJECTION
		if behavioralResult.BehavioralPrompt != "" {
			msgHistory = a.injectBehavioralDirectives(msgHistory, behavioralResult.BehavioralPrompt, behavioralResult.Metadata)
			logging.Info("Behavioral Directives Injected Successfully",
				"directive_length", len(behavioralResult.BehavioralPrompt),
				"enhancement_type", behavioral.GetEnhancementName(behavioralResult.Enhancement))
		} else {
			logging.ErrorPersist("Critical Error: No behavioral prompt generated - framework may be misconfigured")
		}
	}

	for {
		// Check for cancellation before each iteration
		select {
		case <-ctx.Done():
			return a.err(ctx.Err())
		default:
			// Continue processing
		}
		agentMessage, toolResults, err := a.streamAndHandleEvents(ctx, sessionID, msgHistory)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				agentMessage.AddFinish(message.FinishReasonCanceled)
				a.messages.Update(context.Background(), agentMessage)
				return a.err(ErrRequestCancelled)
			}
			return a.err(fmt.Errorf("failed to process events: %w", err))
		}
		if cfg.Debug {
			seqId := (len(msgHistory) + 1) / 2
			toolResultFilepath := logging.WriteToolResultsJson(sessionID, seqId, toolResults)
			logging.Info("Result", "message", agentMessage.FinishReason(), "toolResults", "{}", "filepath", toolResultFilepath)
		} else {
			logging.Info("Result", "message", agentMessage.FinishReason(), "toolResults", toolResults)
		}
		if (agentMessage.FinishReason() == message.FinishReasonToolUse) && toolResults != nil {
			// We are not done, we need to respond with the tool response
			msgHistory = append(msgHistory, agentMessage, *toolResults)
			continue
		}
		return AgentEvent{
			Type:    AgentEventTypeResponse,
			Message: agentMessage,
			Done:    true,
		}
	}
}

func (a *agent) createUserMessage(ctx context.Context, sessionID, content string, attachmentParts []message.ContentPart) (message.Message, error) {
	parts := []message.ContentPart{message.TextContent{Text: content}}
	parts = append(parts, attachmentParts...)
	return a.messages.Create(ctx, sessionID, message.CreateMessageParams{
		Role:  message.User,
		Parts: parts,
	})
}

func (a *agent) streamAndHandleEvents(ctx context.Context, sessionID string, msgHistory []message.Message) (message.Message, *message.Message, error) {
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, sessionID)
	eventChan := a.provider.StreamResponse(ctx, msgHistory, a.tools)

	assistantMsg, err := a.messages.Create(ctx, sessionID, message.CreateMessageParams{
		Role:  message.Assistant,
		Parts: []message.ContentPart{},
		Model: a.provider.Model().ID,
	})
	if err != nil {
		return assistantMsg, nil, fmt.Errorf("failed to create assistant message: %w", err)
	}

	// Add the session and message ID into the context if needed by tools.
	ctx = context.WithValue(ctx, tools.MessageIDContextKey, assistantMsg.ID)

	// Process each event in the stream.
	for event := range eventChan {
		if processErr := a.processEvent(ctx, sessionID, &assistantMsg, event); processErr != nil {
			a.finishMessage(ctx, &assistantMsg, message.FinishReasonCanceled)
			return assistantMsg, nil, processErr
		}
		if ctx.Err() != nil {
			a.finishMessage(context.Background(), &assistantMsg, message.FinishReasonCanceled)
			return assistantMsg, nil, ctx.Err()
		}
	}

	toolResults := make([]message.ToolResult, len(assistantMsg.ToolCalls()))
	toolCalls := assistantMsg.ToolCalls()
	for i, toolCall := range toolCalls {
		select {
		case <-ctx.Done():
			a.finishMessage(context.Background(), &assistantMsg, message.FinishReasonCanceled)
			// Make all future tool calls cancelled
			for j := i; j < len(toolCalls); j++ {
				toolResults[j] = message.ToolResult{
					ToolCallID: toolCalls[j].ID,
					Content:    "Tool execution canceled by user",
					IsError:    true,
				}
			}
			goto out
		default:
			// Continue processing
			var tool tools.BaseTool
			for _, availableTool := range a.tools {
				if availableTool.Info().Name == toolCall.Name {
					tool = availableTool
					break
				}
				// Monkey patch for Copilot Sonnet-4 tool repetition obfuscation
				// if strings.HasPrefix(toolCall.Name, availableTool.Info().Name) &&
				// 	strings.HasPrefix(toolCall.Name, availableTool.Info().Name+availableTool.Info().Name) {
				// 	tool = availableTool
				// 	break
				// }
			}

			// Tool not found
			if tool == nil {
				toolResults[i] = message.ToolResult{
					ToolCallID: toolCall.ID,
					Content:    fmt.Sprintf("Tool not found: %s", toolCall.Name),
					IsError:    true,
				}
				continue
			}
			toolResult, toolErr := tool.Run(ctx, tools.ToolCall{
				ID:    toolCall.ID,
				Name:  toolCall.Name,
				Input: toolCall.Input,
			})
			if toolErr != nil {
				if errors.Is(toolErr, permission.ErrorPermissionDenied) {
					toolResults[i] = message.ToolResult{
						ToolCallID: toolCall.ID,
						Content:    "Permission denied",
						IsError:    true,
					}
					for j := i + 1; j < len(toolCalls); j++ {
						toolResults[j] = message.ToolResult{
							ToolCallID: toolCalls[j].ID,
							Content:    "Tool execution canceled by user",
							IsError:    true,
						}
					}
					a.finishMessage(ctx, &assistantMsg, message.FinishReasonPermissionDenied)
					break
				}
			}
			toolResults[i] = message.ToolResult{
				ToolCallID: toolCall.ID,
				Content:    toolResult.Content,
				Metadata:   toolResult.Metadata,
				IsError:    toolResult.IsError,
			}
		}
	}
out:
	if len(toolResults) == 0 {
		return assistantMsg, nil, nil
	}
	parts := make([]message.ContentPart, 0)
	for _, tr := range toolResults {
		parts = append(parts, tr)
	}
	msg, err := a.messages.Create(context.Background(), assistantMsg.SessionID, message.CreateMessageParams{
		Role:  message.Tool,
		Parts: parts,
	})
	if err != nil {
		return assistantMsg, nil, fmt.Errorf("failed to create cancelled tool message: %w", err)
	}

	return assistantMsg, &msg, err
}

func (a *agent) finishMessage(ctx context.Context, msg *message.Message, finishReson message.FinishReason) {
	msg.AddFinish(finishReson)
	_ = a.messages.Update(ctx, *msg)
}

func (a *agent) processEvent(ctx context.Context, sessionID string, assistantMsg *message.Message, event provider.ProviderEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue processing.
	}

	switch event.Type {
	case provider.EventThinkingDelta:
		assistantMsg.AppendReasoningContent(event.Content)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventContentDelta:
		assistantMsg.AppendContent(event.Content)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventToolUseStart:
		assistantMsg.AddToolCall(*event.ToolCall)
		return a.messages.Update(ctx, *assistantMsg)
	// TODO: see how to handle this
	// case provider.EventToolUseDelta:
	// 	tm := time.Unix(assistantMsg.UpdatedAt, 0)
	// 	assistantMsg.AppendToolCallInput(event.ToolCall.ID, event.ToolCall.Input)
	// 	if time.Since(tm) > 1000*time.Millisecond {
	// 		err := a.messages.Update(ctx, *assistantMsg)
	// 		assistantMsg.UpdatedAt = time.Now().Unix()
	// 		return err
	// 	}
	case provider.EventToolUseStop:
		assistantMsg.FinishToolCall(event.ToolCall.ID)
		return a.messages.Update(ctx, *assistantMsg)
	case provider.EventError:
		if errors.Is(event.Error, context.Canceled) {
			logging.InfoPersist(fmt.Sprintf("Event processing canceled for session: %s", sessionID))
			return context.Canceled
		}
		logging.ErrorPersist(event.Error.Error())
		return event.Error
	case provider.EventComplete:
		assistantMsg.SetToolCalls(event.Response.ToolCalls)
		assistantMsg.AddFinish(event.Response.FinishReason)
		if err := a.messages.Update(ctx, *assistantMsg); err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
		return a.TrackUsage(ctx, sessionID, a.provider.Model(), event.Response.Usage)
	}

	return nil
}

func (a *agent) TrackUsage(ctx context.Context, sessionID string, model models.Model, usage provider.TokenUsage) error {
	sess, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	cost := model.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
		model.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens) +
		model.CostPer1MIn/1e6*float64(usage.InputTokens) +
		model.CostPer1MOut/1e6*float64(usage.OutputTokens)

	sess.Cost += cost
	sess.CompletionTokens = usage.OutputTokens + usage.CacheReadTokens
	sess.PromptTokens = usage.InputTokens + usage.CacheCreationTokens

	_, err = a.sessions.Save(ctx, sess)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (a *agent) Update(agentName config.AgentName, modelID models.ModelID) (models.Model, error) {
	if a.IsBusy() {
		return models.Model{}, fmt.Errorf("cannot change model while processing requests")
	}

	if err := config.UpdateAgentModel(agentName, modelID); err != nil {
		return models.Model{}, fmt.Errorf("failed to update config: %w", err)
	}

	provider, err := createAgentProvider(agentName)
	if err != nil {
		return models.Model{}, fmt.Errorf("failed to create provider for model %s: %w", modelID, err)
	}

	a.provider = provider

	return a.provider.Model(), nil
}

func (a *agent) Summarize(ctx context.Context, sessionID string) error {
	if a.summarizeProvider == nil {
		return fmt.Errorf("summarize provider not available")
	}

	// Check if session is busy
	if a.IsSessionBusy(sessionID) {
		return ErrSessionBusy
	}

	// Create a new context with cancellation
	summarizeCtx, cancel := context.WithCancel(ctx)

	// Store the cancel function in activeRequests to allow cancellation
	a.activeRequests.Store(sessionID+"-summarize", cancel)

	go func() {
		defer a.activeRequests.Delete(sessionID + "-summarize")
		defer cancel()
		event := AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Starting summarization...",
		}

		a.Publish(pubsub.CreatedEvent, event)
		// Get all messages from the session
		msgs, err := a.messages.List(summarizeCtx, sessionID)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to list messages: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		summarizeCtx = context.WithValue(summarizeCtx, tools.SessionIDContextKey, sessionID)

		if len(msgs) == 0 {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("no messages to summarize"),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}

		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Analyzing conversation...",
		}
		a.Publish(pubsub.CreatedEvent, event)

		// Add a system message to guide the summarization
		summarizePrompt := "Provide a detailed but concise summary of our conversation above. Focus on information that would be helpful for continuing the conversation, including what we did, what we're doing, which files we're working on, and what we're going to do next."

		// Create a new message with the summarize prompt
		promptMsg := message.Message{
			Role:  message.User,
			Parts: []message.ContentPart{message.TextContent{Text: summarizePrompt}},
		}

		// Append the prompt to the messages
		msgsWithPrompt := append(msgs, promptMsg)

		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Generating summary...",
		}

		a.Publish(pubsub.CreatedEvent, event)

		// Send the messages to the summarize provider
		response, err := a.summarizeProvider.SendMessages(
			summarizeCtx,
			msgsWithPrompt,
			make([]tools.BaseTool, 0),
		)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to summarize: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}

		summary := strings.TrimSpace(response.Content)
		if summary == "" {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("empty summary returned"),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		event = AgentEvent{
			Type:     AgentEventTypeSummarize,
			Progress: "Creating new session...",
		}

		a.Publish(pubsub.CreatedEvent, event)
		oldSession, err := a.sessions.Get(summarizeCtx, sessionID)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to get session: %w", err),
				Done:  true,
			}

			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		// Create a message in the new session with the summary
		msg, err := a.messages.Create(summarizeCtx, oldSession.ID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.TextContent{Text: summary},
				message.Finish{
					Reason: message.FinishReasonEndTurn,
					Time:   time.Now().Unix(),
				},
			},
			Model: a.summarizeProvider.Model().ID,
		})
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to create summary message: %w", err),
				Done:  true,
			}

			a.Publish(pubsub.CreatedEvent, event)
			return
		}
		oldSession.SummaryMessageID = msg.ID
		oldSession.CompletionTokens = response.Usage.OutputTokens
		oldSession.PromptTokens = 0
		model := a.summarizeProvider.Model()
		usage := response.Usage
		cost := model.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
			model.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens) +
			model.CostPer1MIn/1e6*float64(usage.InputTokens) +
			model.CostPer1MOut/1e6*float64(usage.OutputTokens)
		oldSession.Cost += cost
		_, err = a.sessions.Save(summarizeCtx, oldSession)
		if err != nil {
			event = AgentEvent{
				Type:  AgentEventTypeError,
				Error: fmt.Errorf("failed to save session: %w", err),
				Done:  true,
			}
			a.Publish(pubsub.CreatedEvent, event)
		}

		event = AgentEvent{
			Type:      AgentEventTypeSummarize,
			SessionID: oldSession.ID,
			Progress:  "Summary complete",
			Done:      true,
		}
		a.Publish(pubsub.CreatedEvent, event)
		// Send final success event with the new session ID
	}()

	return nil
}

func createAgentProvider(agentName config.AgentName) (provider.Provider, error) {
	cfg := config.Get()
	agentConfig, ok := cfg.Agents[agentName]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	model, ok := models.SupportedModels[agentConfig.Model]
	if !ok {
		return nil, fmt.Errorf("model %s not supported", agentConfig.Model)
	}

	providerCfg, ok := cfg.Providers[model.Provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", model.Provider)
	}
	if providerCfg.Disabled {
		return nil, fmt.Errorf("provider %s is not enabled", model.Provider)
	}
	maxTokens := model.DefaultMaxTokens
	if agentConfig.MaxTokens > 0 {
		maxTokens = agentConfig.MaxTokens
	}
	opts := []provider.ProviderClientOption{
		provider.WithAPIKey(providerCfg.APIKey),
		provider.WithModel(model),
		provider.WithSystemMessage(prompt.GetAgentPrompt(agentName, model.Provider)),
		provider.WithMaxTokens(maxTokens),
	}
	if model.Provider == models.ProviderOpenAI || model.Provider == models.ProviderLocal && model.CanReason {
		opts = append(
			opts,
			provider.WithOpenAIOptions(
				provider.WithReasoningEffort(agentConfig.ReasoningEffort),
			),
		)
	} else if model.Provider == models.ProviderAnthropic && model.CanReason && agentName == config.AgentCoder {
		opts = append(
			opts,
			provider.WithAnthropicOptions(
				provider.WithAnthropicShouldThinkFn(provider.DefaultShouldThinkFn),
			),
		)
	}
	agentProvider, err := provider.NewProvider(
		model.Provider,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create provider: %v", err)
	}

	return agentProvider, nil
}
