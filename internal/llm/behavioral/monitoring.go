package behavioral

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opencode-ai/opencode/internal/logging"
)

// PerformanceMonitor implements comprehensive performance monitoring for the behavioral framework
type PerformanceMonitor struct {
	enabled       bool
	metricsFile   string
	sessions      map[string]*SessionMetrics
	globalMetrics *GlobalMetrics
	mutex         sync.RWMutex
	startTime     time.Time
}

// GlobalMetrics tracks overall system performance
type GlobalMetrics struct {
	TotalRequests          int                       `json:"total_requests"`
	TotalSessions          int                       `json:"total_sessions"`
	EnhancementUsage       map[EnhancementType]int   `json:"enhancement_usage"`
	AverageProcessingTime  time.Duration             `json:"average_processing_time"`
	AverageQualityScore    float64                   `json:"average_quality_score"`
	SystemUptime           time.Duration             `json:"system_uptime"`
	PerformanceTargetsMet  map[string]bool           `json:"performance_targets_met"`
	QualityImprovementRate float64                   `json:"quality_improvement_rate"`
	UserSatisfactionRate   float64                   `json:"user_satisfaction_rate"`
	LastUpdated            time.Time                 `json:"last_updated"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(enabled bool, dataDir string) *PerformanceMonitor {
	metricsFile := filepath.Join(dataDir, "behavioral_metrics.json")
	
	monitor := &PerformanceMonitor{
		enabled:     enabled,
		metricsFile: metricsFile,
		sessions:    make(map[string]*SessionMetrics),
		globalMetrics: &GlobalMetrics{
			EnhancementUsage:      make(map[EnhancementType]int),
			PerformanceTargetsMet: make(map[string]bool),
		},
		startTime: time.Now(),
	}

	if enabled {
		monitor.loadMetrics()
		go monitor.periodicSave()
	}

	return monitor
}

// RecordRequest records a new request and its processing results
func (pm *PerformanceMonitor) RecordRequest(sessionID string, analysis *RequestAnalysis, result *ProcessingResult) {
	if !pm.enabled {
		return
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Update session metrics
	if pm.sessions[sessionID] == nil {
		pm.sessions[sessionID] = &SessionMetrics{
			SessionID:        sessionID,
			EnhancementUsage: make(map[EnhancementType]int),
			QualityScores:    make([]float64, 0),
			UserSatisfaction: make([]float64, 0),
		}
		pm.globalMetrics.TotalSessions++
	}

	session := pm.sessions[sessionID]
	session.RequestCount++
	session.EnhancementUsage[result.Enhancement]++
	session.QualityScores = append(session.QualityScores, result.QualityMetrics.QualityScore)
	session.LastUpdated = time.Now()

	// Calculate running average processing time
	if session.RequestCount == 1 {
		session.AverageProcessTime = result.ProcessingTime
	} else {
		total := session.AverageProcessTime * time.Duration(session.RequestCount-1)
		session.AverageProcessTime = (total + result.ProcessingTime) / time.Duration(session.RequestCount)
	}

	// Update global metrics
	pm.globalMetrics.TotalRequests++
	pm.globalMetrics.EnhancementUsage[result.Enhancement]++
	pm.globalMetrics.LastUpdated = time.Now()
	pm.globalMetrics.SystemUptime = time.Since(pm.startTime)

	// Calculate global averages
	pm.updateGlobalAverages()

	// Check performance targets
	pm.checkPerformanceTargets(result)

	logging.Debug("Performance Metrics Recorded",
		"session_id", sessionID,
		"enhancement", GetEnhancementName(result.Enhancement),
		"processing_time", result.ProcessingTime,
		"quality_score", result.QualityMetrics.QualityScore,
		"total_requests", pm.globalMetrics.TotalRequests)
}

// updateGlobalAverages calculates global performance averages
func (pm *PerformanceMonitor) updateGlobalAverages() {
	if pm.globalMetrics.TotalRequests == 0 {
		return
	}

	var totalProcessingTime time.Duration
	var totalQualityScore float64
	requestCount := 0

	for _, session := range pm.sessions {
		totalProcessingTime += session.AverageProcessTime * time.Duration(session.RequestCount)
		for _, score := range session.QualityScores {
			totalQualityScore += score
		}
		requestCount += session.RequestCount
	}

	if requestCount > 0 {
		pm.globalMetrics.AverageProcessingTime = totalProcessingTime / time.Duration(requestCount)
		pm.globalMetrics.AverageQualityScore = totalQualityScore / float64(requestCount)
	}
}

// checkPerformanceTargets evaluates if performance targets are being met
func (pm *PerformanceMonitor) checkPerformanceTargets(result *ProcessingResult) {
	targets := GetDefaultTargets()

	// Check response time targets
	if result.Enhancement == StandardProcessing {
		pm.globalMetrics.PerformanceTargetsMet["response_time_standard"] = 
			result.ProcessingTime <= targets.ResponseTimeStandard
	} else {
		pm.globalMetrics.PerformanceTargetsMet["response_time_enhanced"] = 
			result.ProcessingTime <= targets.ResponseTimeEnhanced
	}

	// Check quality targets
	pm.globalMetrics.PerformanceTargetsMet["quality_score"] = 
		result.QualityMetrics.QualityScore >= 0.85

	// Check improvement targets
	pm.globalMetrics.PerformanceTargetsMet["quality_improvement"] = 
		result.QualityMetrics.ExpectedImprovement >= targets.QualityImprovement

	// Calculate improvement rate
	if pm.globalMetrics.TotalRequests > 10 {
		enhancedRequests := 0
		for enhancement, count := range pm.globalMetrics.EnhancementUsage {
			if enhancement != StandardProcessing {
				enhancedRequests += count
			}
		}
		pm.globalMetrics.QualityImprovementRate = float64(enhancedRequests) / float64(pm.globalMetrics.TotalRequests)
	}
}

// GetSessionMetrics returns metrics for a specific session
func (pm *PerformanceMonitor) GetSessionMetrics(sessionID string) *SessionMetrics {
	if !pm.enabled {
		return nil
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.sessions[sessionID]
}

// GetGlobalMetrics returns overall system metrics
func (pm *PerformanceMonitor) GetGlobalMetrics() *GlobalMetrics {
	if !pm.enabled {
		return nil
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// Update uptime
	pm.globalMetrics.SystemUptime = time.Since(pm.startTime)
	
	return pm.globalMetrics
}

// GetPerformanceReport generates a comprehensive performance report
func (pm *PerformanceMonitor) GetPerformanceReport() map[string]interface{} {
	if !pm.enabled {
		return map[string]interface{}{
			"enabled": false,
			"message": "Performance monitoring is disabled",
		}
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	report := map[string]interface{}{
		"enabled":        true,
		"global_metrics": pm.globalMetrics,
		"session_count":  len(pm.sessions),
		"generated_at":   time.Now(),
	}

	// Add enhancement effectiveness analysis
	enhancementEffectiveness := make(map[string]interface{})
	for enhancement, count := range pm.globalMetrics.EnhancementUsage {
		if count > 0 {
			percentage := float64(count) / float64(pm.globalMetrics.TotalRequests) * 100
			enhancementEffectiveness[GetEnhancementName(enhancement)] = map[string]interface{}{
				"usage_count":  count,
				"usage_percent": percentage,
			}
		}
	}
	report["enhancement_effectiveness"] = enhancementEffectiveness

	// Add performance target status
	targetsMet := 0
	totalTargets := len(pm.globalMetrics.PerformanceTargetsMet)
	for _, met := range pm.globalMetrics.PerformanceTargetsMet {
		if met {
			targetsMet++
		}
	}
	
	if totalTargets > 0 {
		report["performance_target_compliance"] = map[string]interface{}{
			"targets_met":     targetsMet,
			"total_targets":   totalTargets,
			"compliance_rate": float64(targetsMet) / float64(totalTargets) * 100,
		}
	}

	return report
}

// saveMetrics saves metrics to disk
func (pm *PerformanceMonitor) saveMetrics() error {
	if !pm.enabled {
		return nil
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	data := map[string]interface{}{
		"global_metrics": pm.globalMetrics,
		"sessions":       pm.sessions,
		"saved_at":       time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(pm.metricsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	if err := os.WriteFile(pm.metricsFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// loadMetrics loads metrics from disk
func (pm *PerformanceMonitor) loadMetrics() {
	if !pm.enabled {
		return
	}

	data, err := os.ReadFile(pm.metricsFile)
	if err != nil {
		// File doesn't exist or can't be read, start fresh
		return
	}

	var savedData map[string]interface{}
	if err := json.Unmarshal(data, &savedData); err != nil {
		logging.ErrorPersist(fmt.Sprintf("Failed to unmarshal metrics: %v", err))
		return
	}

	// Load global metrics if available
	if globalData, ok := savedData["global_metrics"]; ok {
		if globalBytes, err := json.Marshal(globalData); err == nil {
			json.Unmarshal(globalBytes, pm.globalMetrics)
		}
	}

	// Load session metrics if available
	if sessionData, ok := savedData["sessions"]; ok {
		if sessionBytes, err := json.Marshal(sessionData); err == nil {
			json.Unmarshal(sessionBytes, &pm.sessions)
		}
	}

	logging.Info("Performance metrics loaded", "file", pm.metricsFile)
}

// periodicSave saves metrics periodically
func (pm *PerformanceMonitor) periodicSave() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := pm.saveMetrics(); err != nil {
			logging.ErrorPersist(fmt.Sprintf("Failed to save metrics: %v", err))
		}
	}
}

// Shutdown gracefully shuts down the performance monitor
func (pm *PerformanceMonitor) Shutdown() error {
	if !pm.enabled {
		return nil
	}

	return pm.saveMetrics()
}
