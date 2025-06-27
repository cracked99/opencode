package behavioral

import (
	"strings"
	"time"
)

// ProcessingResult represents the result of behavioral framework processing
type ProcessingResult struct {
	Enhancement      EnhancementType    `json:"enhancement"`
	SystemPrompt     string             `json:"system_prompt"`
	BehavioralPrompt string             `json:"behavioral_prompt"`
	ProcessingTime   time.Duration      `json:"processing_time"`
	QualityMetrics   QualityMetrics     `json:"quality_metrics"`
	Fallback         bool               `json:"fallback"`
	Metadata         ProcessingMetadata `json:"metadata"`
}

// QualityMetrics tracks quality measurements
type QualityMetrics struct {
	ExpectedImprovement float64 `json:"expected_improvement"`
	ConfidenceLevel     float64 `json:"confidence_level"`
	QualityScore        float64 `json:"quality_score"`
	ConsistencyScore    float64 `json:"consistency_score"`
}

// ProcessingMetadata contains additional processing information
type ProcessingMetadata struct {
	ComplexityLevel    int                `json:"complexity_level"`
	ToolIntensity      int                `json:"tool_intensity"`
	ProblemType        ProblemType        `json:"problem_type"`
	EnhancementReason  string             `json:"enhancement_reason"`
	EstimatedCost      float64            `json:"estimated_cost"`
	PerformanceTargets map[string]float64 `json:"performance_targets"`
}

// PerformanceTracker tracks behavioral framework performance
type PerformanceTracker struct {
	enabled   bool
	metrics   map[string]*SessionMetrics
	startTime time.Time
}

// SessionMetrics tracks metrics for a specific session
type SessionMetrics struct {
	SessionID          string                  `json:"session_id"`
	RequestCount       int                     `json:"request_count"`
	EnhancementUsage   map[EnhancementType]int `json:"enhancement_usage"`
	AverageProcessTime time.Duration           `json:"average_process_time"`
	QualityScores      []float64               `json:"quality_scores"`
	UserSatisfaction   []float64               `json:"user_satisfaction"`
	LastUpdated        time.Time               `json:"last_updated"`
}

// PatternRecognizer handles pattern recognition and problem classification
type PatternRecognizer struct {
	patterns map[ProblemType][]string
}

// EnhancementSelector handles enhancement selection logic
type EnhancementSelector struct {
	thresholds ComplexityThresholds
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(enabled bool) *PerformanceTracker {
	return &PerformanceTracker{
		enabled:   enabled,
		metrics:   make(map[string]*SessionMetrics),
		startTime: time.Now(),
	}
}

// RecordProcessing records processing metrics
func (pt *PerformanceTracker) RecordProcessing(sessionID string, analysis *RequestAnalysis, result *ProcessingResult) {
	if !pt.enabled {
		return
	}

	if pt.metrics[sessionID] == nil {
		pt.metrics[sessionID] = &SessionMetrics{
			SessionID:        sessionID,
			EnhancementUsage: make(map[EnhancementType]int),
			QualityScores:    make([]float64, 0),
			UserSatisfaction: make([]float64, 0),
		}
	}

	metrics := pt.metrics[sessionID]
	metrics.RequestCount++
	metrics.EnhancementUsage[result.Enhancement]++
	metrics.QualityScores = append(metrics.QualityScores, result.QualityMetrics.QualityScore)
	metrics.LastUpdated = time.Now()

	// Calculate average processing time
	if metrics.RequestCount == 1 {
		metrics.AverageProcessTime = result.ProcessingTime
	} else {
		// Running average
		total := metrics.AverageProcessTime * time.Duration(metrics.RequestCount-1)
		metrics.AverageProcessTime = (total + result.ProcessingTime) / time.Duration(metrics.RequestCount)
	}
}

// GetSessionMetrics returns metrics for a specific session
func (pt *PerformanceTracker) GetSessionMetrics(sessionID string) *SessionMetrics {
	return pt.metrics[sessionID]
}

// GetOverallMetrics returns overall performance metrics
func (pt *PerformanceTracker) GetOverallMetrics() map[string]interface{} {
	if !pt.enabled {
		return nil
	}

	totalRequests := 0
	totalEnhancements := make(map[EnhancementType]int)
	var totalProcessTime time.Duration
	var totalQualityScore float64
	sessionCount := len(pt.metrics)

	for _, metrics := range pt.metrics {
		totalRequests += metrics.RequestCount
		totalProcessTime += metrics.AverageProcessTime * time.Duration(metrics.RequestCount)

		for enhancement, count := range metrics.EnhancementUsage {
			totalEnhancements[enhancement] += count
		}

		for _, score := range metrics.QualityScores {
			totalQualityScore += score
		}
	}

	avgProcessTime := time.Duration(0)
	avgQualityScore := 0.0

	if totalRequests > 0 {
		avgProcessTime = totalProcessTime / time.Duration(totalRequests)
		avgQualityScore = totalQualityScore / float64(totalRequests)
	}

	return map[string]interface{}{
		"total_requests":        totalRequests,
		"total_sessions":        sessionCount,
		"enhancement_usage":     totalEnhancements,
		"average_process_time":  avgProcessTime,
		"average_quality_score": avgQualityScore,
		"uptime":                time.Since(pt.startTime),
	}
}

// NewPatternRecognizer creates a new pattern recognizer
func NewPatternRecognizer() *PatternRecognizer {
	patterns := map[ProblemType][]string{
		TechnicalIssue: {
			"bug", "error", "exception", "crash", "fail", "broken", "issue", "problem",
			"not working", "doesn't work", "won't run", "can't execute",
		},
		ProcessImprovement: {
			"optimize", "improve", "enhance", "better", "faster", "efficient", "performance",
			"refactor", "clean up", "streamline", "upgrade",
		},
		DecisionMaking: {
			"choose", "decide", "select", "option", "alternative", "which", "should I",
			"recommend", "suggest", "best", "compare",
		},
		Troubleshooting: {
			"debug", "diagnose", "investigate", "find", "locate", "identify", "trace",
			"why", "how", "what's wrong", "what happened",
		},
		CodeGeneration: {
			"create", "generate", "write", "build", "implement", "develop", "code",
			"function", "class", "method", "script", "program",
		},
		Analysis: {
			"analyze", "examine", "review", "assess", "evaluate", "study", "understand",
			"explain", "describe", "what does", "how does",
		},
	}

	return &PatternRecognizer{
		patterns: patterns,
	}
}

// ClassifyProblem classifies the problem type based on content
func (pr *PatternRecognizer) ClassifyProblem(content string) ProblemType {
	contentLower := strings.ToLower(content)

	scores := make(map[ProblemType]int)

	for problemType, keywords := range pr.patterns {
		for _, keyword := range keywords {
			if strings.Contains(contentLower, keyword) {
				scores[problemType]++
			}
		}
	}

	// Find the problem type with the highest score
	maxScore := 0
	resultType := General

	for problemType, score := range scores {
		if score > maxScore {
			maxScore = score
			resultType = problemType
		}
	}

	return resultType
}

// NewEnhancementSelector creates a new enhancement selector
func NewEnhancementSelector(thresholds ComplexityThresholds) *EnhancementSelector {
	return &EnhancementSelector{
		thresholds: thresholds,
	}
}

// SelectEnhancement selects the appropriate enhancement based on analysis
func (es *EnhancementSelector) SelectEnhancement(analysis *RequestAnalysis) EnhancementType {
	complexity := analysis.ComplexityLevel
	toolIntensity := analysis.ToolIntensity

	// Apply decision tree logic from the framework
	if complexity <= es.thresholds.StandardMax && toolIntensity <= 1 {
		return StandardProcessing
	}

	if complexity >= es.thresholds.CombinedMin {
		return CombinedEnhancement
	}

	if complexity >= es.thresholds.MetaPromptingMin {
		return MetaPromptingEnhancement
	}

	if complexity >= es.thresholds.ReActMin && complexity <= es.thresholds.ReActMax {
		return ReActEnhancement
	}

	if toolIntensity >= 3 {
		return ACIOptimization
	}

	if complexity >= es.thresholds.ReflexionMin && complexity <= es.thresholds.ReflexionMax {
		return ReflexionEnhancement
	}

	return StandardProcessing
}

// GetEnhancementName returns the human-readable name of an enhancement
func GetEnhancementName(enhancement EnhancementType) string {
	switch enhancement {
	case StandardProcessing:
		return "Standard Processing"
	case ReflexionEnhancement:
		return "Reflexion Enhancement"
	case ReActEnhancement:
		return "ReAct Enhancement"
	case ACIOptimization:
		return "ACI Optimization"
	case MetaPromptingEnhancement:
		return "Meta-Prompting Enhancement"
	case CombinedEnhancement:
		return "Combined Enhancement"
	default:
		return "Unknown Enhancement"
	}
}
