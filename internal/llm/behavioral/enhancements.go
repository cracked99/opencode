package behavioral

import (
	"context"
)

// applyEnhancement applies the selected enhancement to the request
func (bf *BehavioralFramework) applyEnhancement(ctx context.Context, sessionID, content string, analysis *RequestAnalysis, enhancement EnhancementType) (*ProcessingResult, error) {
	switch enhancement {
	case StandardProcessing:
		return bf.applyStandardProcessing(content, analysis)
	case ReflexionEnhancement:
		return bf.applyReflexionEnhancement(content, analysis)
	case ReActEnhancement:
		return bf.applyReActEnhancement(content, analysis)
	case ACIOptimization:
		return bf.applyACIOptimization(content, analysis)
	case MetaPromptingEnhancement:
		return bf.applyMetaPromptingEnhancement(content, analysis)
	case CombinedEnhancement:
		return bf.applyCombinedEnhancement(content, analysis)
	default:
		return bf.applyStandardProcessing(content, analysis)
	}
}

// applyStandardProcessing applies standard processing (1x cost, 60 seconds)
func (bf *BehavioralFramework) applyStandardProcessing(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildStandardPrompt(analysis)
	behavioralPrompt := bf.buildCoreBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      StandardProcessing,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0, // Will be set by caller
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.0,
			ConfidenceLevel:     0.85,
			QualityScore:        0.85,
			ConsistencyScore:    0.80,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "Standard processing for low-complexity requests",
			EstimatedCost:     1.0,
			PerformanceTargets: map[string]float64{
				"response_time": 60.0,
				"quality":       85.0,
			},
		},
	}, nil
}

// applyReflexionEnhancement applies reflexion enhancement (2-3x cost, 120-180 seconds)
func (bf *BehavioralFramework) applyReflexionEnhancement(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildEnhancedPrompt(analysis, "reflexion")
	behavioralPrompt := bf.buildReflexionBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      ReflexionEnhancement,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0,
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.10, // 5-15% improvement
			ConfidenceLevel:     0.90,
			QualityScore:        0.90,
			ConsistencyScore:    0.85,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "Reflexion enhancement for quality improvement",
			EstimatedCost:     2.5,
			PerformanceTargets: map[string]float64{
				"response_time":       150.0,
				"quality_improvement": 10.0,
				"accuracy":            90.0,
			},
		},
	}, nil
}

// applyReActEnhancement applies ReAct enhancement (2-4x cost, 180-240 seconds)
func (bf *BehavioralFramework) applyReActEnhancement(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildEnhancedPrompt(analysis, "react")
	behavioralPrompt := bf.buildReActBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      ReActEnhancement,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0,
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.16, // 12-20% improvement
			ConfidenceLevel:     0.85,
			QualityScore:        0.88,
			ConsistencyScore:    0.85,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "ReAct enhancement for systematic reasoning",
			EstimatedCost:     3.0,
			PerformanceTargets: map[string]float64{
				"response_time":          210.0,
				"systematic_improvement": 16.0,
				"problem_resolution":     85.0,
			},
		},
	}, nil
}

// applyACIOptimization applies ACI optimization (Development + cost, 240-300 seconds)
func (bf *BehavioralFramework) applyACIOptimization(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildEnhancedPrompt(analysis, "aci")
	behavioralPrompt := bf.buildACIBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      ACIOptimization,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0,
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.10, // 5-15% improvement
			ConfidenceLevel:     0.80,
			QualityScore:        0.85,
			ConsistencyScore:    0.80,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "ACI optimization for tool-intensive workflows",
			EstimatedCost:     2.0,
			PerformanceTargets: map[string]float64{
				"response_time":         270.0,
				"interface_efficiency":  10.0,
				"workflow_optimization": 80.0,
			},
		},
	}, nil
}

// applyMetaPromptingEnhancement applies meta-prompting enhancement (2-5x cost, 300-360 seconds)
func (bf *BehavioralFramework) applyMetaPromptingEnhancement(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildEnhancedPrompt(analysis, "meta")
	behavioralPrompt := bf.buildMetaPromptingBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      MetaPromptingEnhancement,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0,
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.13, // 8-18% improvement
			ConfidenceLevel:     0.80,
			QualityScore:        0.88,
			ConsistencyScore:    0.85,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "Meta-prompting for adaptive optimization",
			EstimatedCost:     3.5,
			PerformanceTargets: map[string]float64{
				"response_time":         330.0,
				"adaptive_optimization": 13.0,
				"context_alignment":     85.0,
			},
		},
	}, nil
}

// applyCombinedEnhancement applies combined enhancement (5-15x cost, 600-900 seconds)
func (bf *BehavioralFramework) applyCombinedEnhancement(content string, analysis *RequestAnalysis) (*ProcessingResult, error) {
	systemPrompt := bf.buildEnhancedPrompt(analysis, "combined")
	behavioralPrompt := bf.buildCombinedBehavioralDirectives()

	return &ProcessingResult{
		Enhancement:      CombinedEnhancement,
		SystemPrompt:     systemPrompt,
		BehavioralPrompt: behavioralPrompt,
		ProcessingTime:   0,
		QualityMetrics: QualityMetrics{
			ExpectedImprovement: 0.47, // 30-65% improvement
			ConfidenceLevel:     0.85,
			QualityScore:        0.95,
			ConsistencyScore:    0.90,
		},
		Metadata: ProcessingMetadata{
			ComplexityLevel:   analysis.ComplexityLevel,
			ToolIntensity:     analysis.ToolIntensity,
			ProblemType:       analysis.ProblemType,
			EnhancementReason: "Combined enhancement for maximum effectiveness",
			EstimatedCost:     10.0,
			PerformanceTargets: map[string]float64{
				"response_time":              750.0,
				"combined_improvement":       47.0,
				"comprehensive_coordination": 85.0,
			},
		},
	}, nil
}

// buildStandardPrompt builds the standard system prompt
func (bf *BehavioralFramework) buildStandardPrompt(analysis *RequestAnalysis) string {
	base := `You are OpenCode, an interactive CLI tool that helps users with software engineering tasks. 
Use the instructions below and the tools available to you to assist the user systematically and efficiently.

IMPORTANT: Keep your responses concise and focused. Answer directly without unnecessary elaboration.`

	// Add problem-specific guidance
	switch analysis.ProblemType {
	case TechnicalIssue:
		base += "\n\nFocus on identifying the root cause and providing a clear solution."
	case CodeGeneration:
		base += "\n\nGenerate clean, well-documented code following best practices."
	case Analysis:
		base += "\n\nProvide thorough analysis with clear explanations and examples."
	}

	return base
}

// buildEnhancedPrompt builds enhanced system prompts for specific enhancements
func (bf *BehavioralFramework) buildEnhancedPrompt(analysis *RequestAnalysis, enhancementType string) string {
	base := bf.buildStandardPrompt(analysis)

	switch enhancementType {
	case "reflexion":
		base += "\n\nENHANCEMENT: Apply reflexion techniques - generate initial response, evaluate it systematically, then provide improved version."
	case "react":
		base += "\n\nENHANCEMENT: Use systematic reasoning - think through the problem step by step, take actions, observe results, and adapt approach."
	case "aci":
		base += "\n\nENHANCEMENT: Optimize tool usage and workflow efficiency for maximum productivity."
	case "meta":
		base += "\n\nENHANCEMENT: Apply adaptive optimization - customize reasoning approach based on problem context and requirements."
	case "combined":
		base += "\n\nENHANCEMENT: Apply comprehensive enhancement with all available techniques for maximum effectiveness."
	}

	return base
}
