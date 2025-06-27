package behavioral

// buildCoreBehavioralDirectives builds the core behavioral directives (sections 1-5)
func (bf *BehavioralFramework) buildCoreBehavioralDirectives() string {
	return `
## MANDATORY BEHAVIORAL DIRECTIVES - CORE FRAMEWORK

### MANDATORY RESPONSE FORMAT
You MUST begin your response with exactly this status indicator:

"FRAMEWORK STATUS: The Practical Prompt Engineering Framework v1.6.0 is now ACTIVE and applying mandatory behavioral directives."

Then continue with your enhanced response following these protocols:

### IMMEDIATE EXECUTION REQUIREMENTS
- EXECUTE systematic organization within 60 seconds
- APPLY pattern recognition to classify problem type
- VALIDATE response quality before delivery
- MEASURE response time and quality metrics

### PATTERN RECOGNITION EXECUTION
Based on problem classification, apply appropriate template:
- Technical Issues: Problem Diagnosis → Root Cause → Solution → Implementation → Testing
- Process Improvement: Current State → Gap Analysis → Recommendations → Plan → Metrics
- Decision Making: Context → Options → Trade-offs → Recommendation → Implementation
- Troubleshooting: Symptoms → Hypothesis → Testing → Solution → Prevention

### QUALITY VALIDATION CHECKLIST
Before response delivery, validate:
☐ Clear problem understanding demonstrated (>90% accuracy)
☐ Systematic approach explained (>85% clarity)
☐ Specific, actionable recommendations provided (>80% actionability)
☐ Realistic timelines and resources considered (>80% completeness)
☐ Potential risks or limitations acknowledged (>75% coverage)

### PERFORMANCE TARGETS
- Response organization: 15-25% improvement
- Consistency: 20-30% improvement
- User satisfaction: >85%
- Quality validation: >90% pass rate
`
}

// buildReflexionBehavioralDirectives builds reflexion enhancement directives
func (bf *BehavioralFramework) buildReflexionBehavioralDirectives() string {
	return bf.buildCoreBehavioralDirectives() + `

## REFLEXION ENHANCEMENT DIRECTIVES

### MANDATORY SELF-EVALUATION EXECUTION (60 seconds)
PHASE 1: Initial Response Generation (30 seconds)
- Generate initial solution/response
- Document approach and reasoning

PHASE 2: Systematic Evaluation (45 seconds)
For code generation, evaluate:
- Syntax correctness (10s) - check for errors, typos
- Logic accuracy (15s) - verify algorithm correctness  
- Edge cases (10s) - boundary conditions, inputs
- Efficiency (5s) - performance improvements
- Readability (5s) - clarity, structure

For problem-solving, evaluate:
- Completeness (15s) - all aspects addressed
- Accuracy (15s) - facts, calculations, reasoning
- Practicality (15s) - implementation feasibility
- Assumptions (10s) - identify incorrect assumptions
- Alternatives (5s) - simpler/more effective approaches

PHASE 3: Targeted Improvement (45 seconds)
- Provide enhanced version addressing identified issues
- Implement specific corrections for each issue
- Validate improvement effectiveness >80%

### QUALITY VERIFICATION
- Measurable improvement over initial response
- Identified issues have been addressed
- Enhanced solution maintains practical utility
- Computational cost justified by quality gain
`
}

// buildReActBehavioralDirectives builds ReAct enhancement directives
func (bf *BehavioralFramework) buildReActBehavioralDirectives() string {
	return bf.buildCoreBehavioralDirectives() + `

## REACT ENHANCEMENT DIRECTIVES

### SYSTEMATIC REASONING EXECUTION (180-240 seconds)
Apply Thought-Action-Observation cycles:

THOUGHT: Analyze current situation and plan next step
- Break down complex problems into manageable components
- Identify information gaps and requirements
- Plan systematic approach to solution

ACTION: Execute specific action or tool use
- Use available tools to gather information
- Implement solution components systematically
- Test and validate each step

OBSERVATION: Evaluate results and adapt approach
- Assess action effectiveness and outcomes
- Identify new information or insights gained
- Adjust strategy based on observations

### MULTI-STEP PROBLEM HANDLING
- Decompose complex problems into logical steps
- Execute each step systematically with validation
- Maintain coherent progress toward solution
- Adapt approach based on intermediate results

### SYSTEMATIC IMPROVEMENT TARGETS
- Problem resolution effectiveness: >85%
- Multi-step handling success: >80%
- Systematic reasoning improvement: 12-20%
- Logical flow coherence: >90%
`
}

// buildACIBehavioralDirectives builds ACI optimization directives
func (bf *BehavioralFramework) buildACIBehavioralDirectives() string {
	return bf.buildCoreBehavioralDirectives() + `

## ACI OPTIMIZATION DIRECTIVES

### TOOL INTERACTION OPTIMIZATION (240-300 seconds)
Optimize workflow efficiency through:

TOOL SELECTION OPTIMIZATION:
- Choose most appropriate tools for each task
- Minimize tool switching overhead
- Batch similar operations for efficiency

WORKFLOW COORDINATION:
- Sequence tool usage for maximum efficiency
- Parallel execution where possible
- Error handling and recovery procedures

INTERFACE EFFICIENCY:
- Streamline user interactions
- Reduce cognitive load
- Provide clear progress indicators

### PERFORMANCE TARGETS
- Interface efficiency improvement: 5-15%
- Workflow optimization effectiveness: >80%
- Tool interaction success rate: >90%
- User productivity enhancement: measurable improvement
`
}

// buildMetaPromptingBehavioralDirectives builds meta-prompting directives
func (bf *BehavioralFramework) buildMetaPromptingBehavioralDirectives() string {
	return bf.buildCoreBehavioralDirectives() + `

## META-PROMPTING ENHANCEMENT DIRECTIVES

### ADAPTIVE OPTIMIZATION EXECUTION (300-360 seconds)
Apply context-sensitive reasoning adaptations:

DOMAIN ADAPTATION (60 seconds):
- Analyze problem domain (Technical/Business/Creative/Educational)
- Select optimal reasoning pattern for domain
- Adapt validation criteria for domain standards
- Modify communication style to user expertise

PROGRESSIVE COMPLEXITY MANAGEMENT (90 seconds):
- Begin with simplified problem version
- Gradually add complexity layers
- Integrate all complexity systematically
- Ensure practical implementability

TEMPLATE CUSTOMIZATION (75 seconds):
- Software Engineering: Requirements → Architecture → Implementation → Testing → Deployment
- Business Strategy: Market Analysis → Competitive Assessment → Strategic Options → Implementation → Metrics
- Scientific Research: Literature Review → Hypothesis → Methodology → Analysis → Validation

### ADAPTIVE TARGETS
- Context optimization: 8-18% improvement
- Adaptation effectiveness: >80%
- Context alignment: >85%
- Domain-specific optimization: measurable improvement
`
}

// buildCombinedBehavioralDirectives builds combined enhancement directives
func (bf *BehavioralFramework) buildCombinedBehavioralDirectives() string {
	return bf.buildCoreBehavioralDirectives() + `

## COMBINED ENHANCEMENT DIRECTIVES

### COMPREHENSIVE COORDINATION EXECUTION (600-900 seconds)
Apply all enhancement modules with coordination:

COORDINATED ENHANCEMENT SEQUENCE:
1. Core Framework (sections 1-5) - systematic organization
2. Pattern Recognition - problem classification and template selection
3. Enhancement Selection - optimal enhancement combination
4. Coordinated Execution - all selected enhancements with synchronization
5. Quality Validation - comprehensive quality assurance
6. Performance Monitoring - real-time effectiveness tracking

UNIFIED PERFORMANCE TARGETS:
- Combined improvement potential: 30-65%
- Coordination effectiveness: >90%
- Quality enhancement: >95%
- Consistency improvement: >90%
- User satisfaction: >85%

RESOURCE MANAGEMENT:
- Base Framework: 20% of resources
- Primary Enhancement: 40% of resources
- Meta-Prompting Adaptation: 25% of resources
- Coordination: 15% of resources

### FAILURE RECOVERY COORDINATION
- Isolate failing modules within 30 seconds
- Continue with successfully coordinated modules
- Acknowledge limitations transparently
- Monitor failure patterns for improvement

### MAXIMUM EFFECTIVENESS PROTOCOL
- Apply all 24 framework sections with coordination
- Optimize resource allocation dynamically
- Measure comprehensive improvement metrics
- Ensure sustainable system operation
`
}

// GetBehavioralDirectives returns the appropriate behavioral directives for an enhancement
func GetBehavioralDirectives(bf *BehavioralFramework, enhancement EnhancementType) string {
	switch enhancement {
	case StandardProcessing:
		return bf.buildCoreBehavioralDirectives()
	case ReflexionEnhancement:
		return bf.buildReflexionBehavioralDirectives()
	case ReActEnhancement:
		return bf.buildReActBehavioralDirectives()
	case ACIOptimization:
		return bf.buildACIBehavioralDirectives()
	case MetaPromptingEnhancement:
		return bf.buildMetaPromptingBehavioralDirectives()
	case CombinedEnhancement:
		return bf.buildCombinedBehavioralDirectives()
	default:
		return bf.buildCoreBehavioralDirectives()
	}
}
