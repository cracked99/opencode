package behavioral

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opencode-ai/opencode/internal/logging"
)

// Enhancement types based on the Practical Prompt Engineering Framework v1.6.0
type EnhancementType int

const (
	StandardProcessing EnhancementType = iota
	ReflexionEnhancement
	ReActEnhancement
	ACIOptimization
	MetaPromptingEnhancement
	CombinedEnhancement
)

// BehavioralConfig represents configuration for behavioral framework
type BehavioralConfig struct {
	Enabled                bool                     `json:"enabled"`
	EnabledEnhancements    map[EnhancementType]bool `json:"enabled_enhancements"`
	ComplexityThresholds   ComplexityThresholds     `json:"complexity_thresholds"`
	PerformanceTargets     PerformanceTargets       `json:"performance_targets"`
	MaxProcessingTime      time.Duration            `json:"max_processing_time"`
	EnablePerformanceTrack bool                     `json:"enable_performance_tracking"`
}

type ComplexityThresholds struct {
	StandardMax      int `json:"standard_max"`       // 1-3
	ReflexionMin     int `json:"reflexion_min"`      // 4
	ReflexionMax     int `json:"reflexion_max"`      // 6
	ReActMin         int `json:"react_min"`          // 6
	ReActMax         int `json:"react_max"`          // 8
	MetaPromptingMin int `json:"meta_prompting_min"` // 7
	CombinedMin      int `json:"combined_min"`       // 8+
}

type PerformanceTargets struct {
	ResponseTimeStandard   time.Duration `json:"response_time_standard"`   // 60s
	ResponseTimeEnhanced   time.Duration `json:"response_time_enhanced"`   // 600s
	QualityImprovement     float64       `json:"quality_improvement"`      // 15-25%
	ConsistencyImprovement float64       `json:"consistency_improvement"`  // 20-30%
	UserSatisfactionTarget float64       `json:"user_satisfaction_target"` // 85%
}

// RequestAnalysis represents the analysis of a user request
type RequestAnalysis struct {
	ComplexityLevel        int             `json:"complexity_level"`
	ToolIntensity          int             `json:"tool_intensity"`
	QualityRequirement     string          `json:"quality_requirement"`
	Keywords               []string        `json:"keywords"`
	ProblemType            ProblemType     `json:"problem_type"`
	EstimatedCost          float64         `json:"estimated_cost"`
	RecommendedEnhancement EnhancementType `json:"recommended_enhancement"`
}

type ProblemType int

const (
	TechnicalIssue ProblemType = iota
	ProcessImprovement
	DecisionMaking
	Troubleshooting
	CodeGeneration
	Analysis
	General
)

// BehavioralFramework implements the Practical Prompt Engineering Framework v1.6.0
type BehavioralFramework struct {
	config              BehavioralConfig
	performanceTracker  *PerformanceTracker
	performanceMonitor  *PerformanceMonitor
	patternRecognizer   *PatternRecognizer
	enhancementSelector *EnhancementSelector
}

// NewBehavioralFramework creates a new behavioral framework instance
func NewBehavioralFramework(cfg BehavioralConfig) *BehavioralFramework {
	return &BehavioralFramework{
		config:              cfg,
		performanceTracker:  NewPerformanceTracker(cfg.EnablePerformanceTrack),
		performanceMonitor:  NewPerformanceMonitor(cfg.EnablePerformanceTrack, ".opencode"),
		patternRecognizer:   NewPatternRecognizer(),
		enhancementSelector: NewEnhancementSelector(cfg.ComplexityThresholds),
	}
}

// ProcessRequest applies the behavioral framework to a user request
func (bf *BehavioralFramework) ProcessRequest(ctx context.Context, sessionID, content string) (*ProcessingResult, error) {
	if !bf.config.Enabled {
		return &ProcessingResult{
			Enhancement:    StandardProcessing,
			SystemPrompt:   getStandardPrompt(),
			ProcessingTime: 0,
			QualityMetrics: QualityMetrics{},
		}, nil
	}

	startTime := time.Now()

	// MANDATORY EXECUTION: Pattern recognition within 20 seconds
	analysis, err := bf.analyzeRequest(content)
	if err != nil {
		logging.ErrorPersist(fmt.Sprintf("Request analysis failed: %v", err))
		return bf.fallbackToStandard(content)
	}

	// MANDATORY EXECUTION: Enhancement selection within 15 seconds
	enhancement := bf.enhancementSelector.SelectEnhancement(analysis)

	// MANDATORY EXECUTION: Apply selected enhancement
	result, err := bf.applyEnhancement(ctx, sessionID, content, analysis, enhancement)
	if err != nil {
		logging.ErrorPersist(fmt.Sprintf("Enhancement application failed: %v", err))
		return bf.fallbackToStandard(content)
	}

	result.ProcessingTime = time.Since(startTime)

	// Track performance if enabled
	if bf.config.EnablePerformanceTrack {
		bf.performanceTracker.RecordProcessing(sessionID, analysis, result)
		bf.performanceMonitor.RecordRequest(sessionID, analysis, result)
	}

	return result, nil
}

// analyzeRequest performs mandatory pattern recognition and complexity analysis
func (bf *BehavioralFramework) analyzeRequest(content string) (*RequestAnalysis, error) {
	analysis := &RequestAnalysis{
		Keywords: extractKeywords(content),
	}

	// Pattern recognition execution
	analysis.ProblemType = bf.patternRecognizer.ClassifyProblem(content)
	analysis.ComplexityLevel = bf.calculateComplexity(content)
	analysis.ToolIntensity = bf.calculateToolIntensity(content)
	analysis.QualityRequirement = bf.determineQualityRequirement(content)

	return analysis, nil
}

// calculateComplexity determines the complexity level (1-10) of the request
func (bf *BehavioralFramework) calculateComplexity(content string) int {
	complexity := 1

	// Length-based complexity
	if len(content) > 500 {
		complexity += 2
	} else if len(content) > 200 {
		complexity += 1
	}

	// Keyword-based complexity indicators
	complexityKeywords := []string{
		"architecture", "system", "design", "integrate", "complex", "multiple",
		"analyze", "optimize", "refactor", "debug", "troubleshoot", "implement",
		"comprehensive", "detailed", "advanced", "sophisticated",
	}

	for _, keyword := range complexityKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			complexity++
		}
	}

	// Multi-step indicators
	if strings.Contains(content, "and") || strings.Contains(content, "then") {
		complexity++
	}

	// Cap at 10
	if complexity > 10 {
		complexity = 10
	}

	return complexity
}

// calculateToolIntensity determines how tool-intensive the request is (0-5)
func (bf *BehavioralFramework) calculateToolIntensity(content string) int {
	intensity := 0

	toolKeywords := []string{
		"file", "code", "edit", "create", "modify", "run", "execute", "test",
		"build", "compile", "deploy", "install", "configure", "setup",
	}

	for _, keyword := range toolKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			intensity++
		}
	}

	// Cap at 5
	if intensity > 5 {
		intensity = 5
	}

	return intensity
}

// determineQualityRequirement assesses the quality requirement level
func (bf *BehavioralFramework) determineQualityRequirement(content string) string {
	highQualityKeywords := []string{
		"production", "critical", "important", "careful", "precise", "accurate",
		"comprehensive", "detailed", "thorough", "professional",
	}

	for _, keyword := range highQualityKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return "high"
		}
	}

	return "standard"
}

// fallbackToStandard provides fallback when framework fails
func (bf *BehavioralFramework) fallbackToStandard(content string) (*ProcessingResult, error) {
	return &ProcessingResult{
		Enhancement:    StandardProcessing,
		SystemPrompt:   getStandardPrompt(),
		ProcessingTime: 0,
		QualityMetrics: QualityMetrics{},
		Fallback:       true,
	}, nil
}

// extractKeywords extracts relevant keywords from content
func extractKeywords(content string) []string {
	// Simple keyword extraction - can be enhanced with NLP
	words := strings.Fields(strings.ToLower(content))
	keywords := make([]string, 0)

	relevantWords := map[string]bool{
		"code": true, "function": true, "class": true, "method": true,
		"bug": true, "error": true, "fix": true, "debug": true,
		"implement": true, "create": true, "build": true, "design": true,
		"optimize": true, "improve": true, "refactor": true, "analyze": true,
	}

	for _, word := range words {
		if relevantWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// GetPerformanceReport returns a comprehensive performance report
func (bf *BehavioralFramework) GetPerformanceReport() map[string]interface{} {
	if bf.performanceMonitor != nil {
		return bf.performanceMonitor.GetPerformanceReport()
	}
	return map[string]interface{}{
		"enabled": false,
		"message": "Performance monitoring not available",
	}
}

// GetGlobalMetrics returns global performance metrics
func (bf *BehavioralFramework) GetGlobalMetrics() *GlobalMetrics {
	if bf.performanceMonitor != nil {
		return bf.performanceMonitor.GetGlobalMetrics()
	}
	return nil
}

// GetSessionMetrics returns metrics for a specific session
func (bf *BehavioralFramework) GetSessionMetrics(sessionID string) *SessionMetrics {
	if bf.performanceMonitor != nil {
		return bf.performanceMonitor.GetSessionMetrics(sessionID)
	}
	return nil
}

// Shutdown gracefully shuts down the behavioral framework
func (bf *BehavioralFramework) Shutdown() error {
	if bf.performanceMonitor != nil {
		return bf.performanceMonitor.Shutdown()
	}
	return nil
}

// getStandardPrompt returns the standard system prompt
func getStandardPrompt() string {
	return `You are OpenCode, an interactive CLI tool that helps users with software engineering tasks.
Use the instructions below and the tools available to you to assist the user systematically and efficiently.`
}
