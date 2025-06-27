package behavioral

import (
	"fmt"
	"strings"
	"time"
)

// DecisionTree implements the enhancement selection decision tree from the framework
type DecisionTree struct {
	thresholds ComplexityThresholds
	targets    PerformanceTargets
}

// NewDecisionTree creates a new decision tree
func NewDecisionTree(thresholds ComplexityThresholds, targets PerformanceTargets) *DecisionTree {
	return &DecisionTree{
		thresholds: thresholds,
		targets:    targets,
	}
}

// SelectEnhancement implements the mandatory enhancement selection decision tree
func (dt *DecisionTree) SelectEnhancement(analysis *RequestAnalysis) (*EnhancementDecision, error) {
	// MANDATORY EXECUTION: Enhancement selection within 15 seconds
	startTime := time.Now()
	
	decision := &EnhancementDecision{
		Analysis:    analysis,
		Timestamp:   startTime,
		Reasoning:   make([]string, 0),
	}

	// Apply decision criteria: Complexity + Tool Intensity + Quality Requirements + Meta-prompting suitability
	complexity := analysis.ComplexityLevel
	toolIntensity := analysis.ToolIntensity
	qualityReq := analysis.QualityRequirement
	
	decision.addReasoning(fmt.Sprintf("Complexity Level: %d, Tool Intensity: %d, Quality: %s", 
		complexity, toolIntensity, qualityReq))

	// DECISION TREE EXECUTION
	
	// IF Complexity 1-3 AND Tool Intensity 0-1 AND Quality Standard
	if complexity >= 1 && complexity <= dt.thresholds.StandardMax && toolIntensity <= 1 && qualityReq == "standard" {
		decision.Enhancement = StandardProcessing
		decision.EstimatedCost = 1.0
		decision.EstimatedTime = dt.targets.ResponseTimeStandard
		decision.ExpectedImprovement = 0.0
		decision.addReasoning("Standard Processing: Low complexity, low tool intensity, standard quality")
		return decision, nil
	}

	// IF Complexity 4-6 AND Quality High
	if complexity >= dt.thresholds.ReflexionMin && complexity <= dt.thresholds.ReflexionMax && qualityReq == "high" {
		decision.Enhancement = ReflexionEnhancement
		decision.EstimatedCost = 2.5 // 2-3x cost
		decision.EstimatedTime = time.Duration(150) * time.Second // 120-180s
		decision.ExpectedImprovement = 0.10 // 5-15%
		decision.addReasoning("Reflexion Enhancement: Medium complexity with high quality requirement")
		return decision, nil
	}

	// IF Complexity 6-8 AND Multi-step Required
	if complexity >= dt.thresholds.ReActMin && complexity <= dt.thresholds.ReActMax && dt.isMultiStepRequired(analysis) {
		decision.Enhancement = ReActEnhancement
		decision.EstimatedCost = 3.0 // 2-4x cost
		decision.EstimatedTime = time.Duration(210) * time.Second // 180-240s
		decision.ExpectedImprovement = 0.16 // 12-20%
		decision.addReasoning("ReAct Enhancement: High complexity requiring systematic reasoning")
		return decision, nil
	}

	// IF Tool Intensity 3-5
	if toolIntensity >= 3 && toolIntensity <= 5 {
		decision.Enhancement = ACIOptimization
		decision.EstimatedCost = 2.0 // Development + cost
		decision.EstimatedTime = time.Duration(270) * time.Second // 240-300s
		decision.ExpectedImprovement = 0.10 // 5-15%
		decision.addReasoning("ACI Optimization: High tool intensity requiring workflow optimization")
		return decision, nil
	}

	// IF Complexity 7+ AND Context Variability High
	if complexity >= dt.thresholds.MetaPromptingMin && dt.hasHighContextVariability(analysis) {
		decision.Enhancement = MetaPromptingEnhancement
		decision.EstimatedCost = 3.5 // 2-5x cost
		decision.EstimatedTime = time.Duration(330) * time.Second // 300-360s
		decision.ExpectedImprovement = 0.13 // 8-18%
		decision.addReasoning("Meta-Prompting Enhancement: High complexity with context variability")
		return decision, nil
	}

	// IF Multiple Enhancement Criteria Met
	if dt.hasMultipleEnhancementCriteria(analysis) {
		decision.Enhancement = CombinedEnhancement
		decision.EstimatedCost = 10.0 // 5-15x cost
		decision.EstimatedTime = time.Duration(750) * time.Second // 600-900s
		decision.ExpectedImprovement = 0.47 // 30-65%
		decision.addReasoning("Combined Enhancement: Multiple criteria met, maximum effectiveness required")
		return decision, nil
	}

	// Default to Standard Processing if no specific criteria met
	decision.Enhancement = StandardProcessing
	decision.EstimatedCost = 1.0
	decision.EstimatedTime = dt.targets.ResponseTimeStandard
	decision.ExpectedImprovement = 0.0
	decision.addReasoning("Default to Standard Processing: No specific enhancement criteria met")
	
	return decision, nil
}

// isMultiStepRequired determines if the request requires multi-step processing
func (dt *DecisionTree) isMultiStepRequired(analysis *RequestAnalysis) bool {
	multiStepKeywords := []string{
		"step", "steps", "process", "workflow", "sequence", "then", "after", "next",
		"first", "second", "finally", "implement", "deploy", "test", "analyze",
		"multiple", "several", "various", "complex", "comprehensive",
	}

	content := strings.Join(analysis.Keywords, " ")
	for _, keyword := range multiStepKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return true
		}
	}

	// High complexity generally requires multi-step approach
	return analysis.ComplexityLevel >= 6
}

// hasHighContextVariability determines if the request has high context variability
func (dt *DecisionTree) hasHighContextVariability(analysis *RequestAnalysis) bool {
	contextKeywords := []string{
		"adaptive", "context-sensitive", "domain-specific", "customization",
		"different", "various", "multiple", "depending", "based on", "according to",
		"specific", "particular", "unique", "specialized", "tailored",
	}

	content := strings.Join(analysis.Keywords, " ")
	for _, keyword := range contextKeywords {
		if strings.Contains(strings.ToLower(content), keyword) {
			return true
		}
	}

	// Multiple problem types suggest context variability
	return analysis.ComplexityLevel >= 7
}

// hasMultipleEnhancementCriteria determines if multiple enhancement criteria are met
func (dt *DecisionTree) hasMultipleEnhancementCriteria(analysis *RequestAnalysis) bool {
	criteriaCount := 0

	// High complexity
	if analysis.ComplexityLevel >= dt.thresholds.CombinedMin {
		criteriaCount++
	}

	// High tool intensity
	if analysis.ToolIntensity >= 3 {
		criteriaCount++
	}

	// High quality requirement
	if analysis.QualityRequirement == "high" {
		criteriaCount++
	}

	// Multi-step requirement
	if dt.isMultiStepRequired(analysis) {
		criteriaCount++
	}

	// Context variability
	if dt.hasHighContextVariability(analysis) {
		criteriaCount++
	}

	// Multiple criteria met (2 or more)
	return criteriaCount >= 2
}

// ValidateDecision validates the enhancement decision against framework requirements
func (dt *DecisionTree) ValidateDecision(decision *EnhancementDecision) error {
	// Validate cost-benefit ratio
	if decision.EstimatedCost > 0 {
		roi := decision.ExpectedImprovement / decision.EstimatedCost
		if roi < 0.02 { // Minimum 2% ROI
			return fmt.Errorf("enhancement ROI too low: %.2f%%, minimum 2%% required", roi*100)
		}
	}

	// Validate time constraints
	maxTime := time.Duration(900) * time.Second // 15 minutes maximum
	if decision.EstimatedTime > maxTime {
		return fmt.Errorf("estimated time %v exceeds maximum %v", decision.EstimatedTime, maxTime)
	}

	// Validate enhancement appropriateness
	if decision.Enhancement == CombinedEnhancement && decision.Analysis.ComplexityLevel < dt.thresholds.CombinedMin {
		return fmt.Errorf("combined enhancement selected for complexity %d, minimum %d required", 
			decision.Analysis.ComplexityLevel, dt.thresholds.CombinedMin)
	}

	return nil
}

// EnhancementDecision represents the result of enhancement selection
type EnhancementDecision struct {
	Enhancement         EnhancementType   `json:"enhancement"`
	Analysis           *RequestAnalysis  `json:"analysis"`
	EstimatedCost      float64          `json:"estimated_cost"`
	EstimatedTime      time.Duration    `json:"estimated_time"`
	ExpectedImprovement float64         `json:"expected_improvement"`
	Reasoning          []string         `json:"reasoning"`
	Timestamp          time.Time        `json:"timestamp"`
	ValidationErrors   []string         `json:"validation_errors,omitempty"`
}

// addReasoning adds a reasoning step to the decision
func (ed *EnhancementDecision) addReasoning(reason string) {
	ed.Reasoning = append(ed.Reasoning, reason)
}

// GetDecisionSummary returns a human-readable summary of the decision
func (ed *EnhancementDecision) GetDecisionSummary() string {
	summary := fmt.Sprintf("Enhancement: %s\n", GetEnhancementName(ed.Enhancement))
	summary += fmt.Sprintf("Estimated Cost: %.1fx baseline\n", ed.EstimatedCost)
	summary += fmt.Sprintf("Estimated Time: %v\n", ed.EstimatedTime)
	summary += fmt.Sprintf("Expected Improvement: %.1f%%\n", ed.ExpectedImprovement*100)
	summary += fmt.Sprintf("Reasoning:\n")
	for _, reason := range ed.Reasoning {
		summary += fmt.Sprintf("  - %s\n", reason)
	}
	return summary
}

// GetDefaultThresholds returns the default complexity thresholds
func GetDefaultThresholds() ComplexityThresholds {
	return ComplexityThresholds{
		StandardMax:      3,
		ReflexionMin:     4,
		ReflexionMax:     6,
		ReActMin:         6,
		ReActMax:         8,
		MetaPromptingMin: 7,
		CombinedMin:      8,
	}
}

// GetDefaultTargets returns the default performance targets
func GetDefaultTargets() PerformanceTargets {
	return PerformanceTargets{
		ResponseTimeStandard:   time.Duration(60) * time.Second,
		ResponseTimeEnhanced:   time.Duration(600) * time.Second,
		QualityImprovement:     0.20, // 20%
		ConsistencyImprovement: 0.25, // 25%
		UserSatisfactionTarget: 0.85, // 85%
	}
}
