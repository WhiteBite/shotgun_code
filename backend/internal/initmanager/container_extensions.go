package initmanager

import (
	"context"
	"shotgun_code/application"
	"shotgun_code/domain"
	contextservice "shotgun_code/internal/context"
)

// LazyServices holds lazy-initialized services
type LazyServices struct {
	// AI Services
	AIService *LazyService[*application.AIService]

	// Context Services (unified service replaces old separate services)
	ContextService  *LazyService[*contextservice.Service]
	ContextBuilder  *LazyService[domain.ContextBuilder]
	ContextAnalyzer *LazyService[domain.ContextAnalyzer]

	// Analysis Services
	TestService           *LazyService[domain.ITestService]
	StaticAnalyzerService *LazyService[domain.IStaticAnalyzerService]
	SBOMService           *LazyService[*application.SBOMService]

	// Build and Verification Services
	BuildService                *LazyService[domain.IBuildService]
	VerificationPipelineService *LazyService[*application.VerificationPipelineService]

	// Export and Report Services
	ExportService *LazyService[*application.ExportService]
	ReportService *LazyService[*application.ReportService]

	// Task Protocol Services
	TaskProtocolService *LazyService[domain.TaskProtocolService]
	ErrorAnalyzer       *LazyService[domain.ErrorAnalyzer]
	CorrectionEngine    *LazyService[domain.CorrectionEngine]

	// Symbol Graph Service
	SymbolGraph *LazyService[*application.SymbolGraphService]

	// Repair Service
	RepairService *LazyService[domain.RepairService]

	// Guardrail Service
	GuardrailService *LazyService[domain.GuardrailService]

	// Taskflow Service
	TaskflowService *LazyService[domain.TaskflowService]

	// UX Metrics Service
	UXMetricsService *LazyService[domain.UXMetricsService]

	// Apply and Diff Services
	ApplyService *LazyService[*application.ApplyService]
	DiffService  *LazyService[*application.DiffService]

	// Router LLM Service
	RouterLLMService *LazyService[*application.RouterLLMService]
}

// NewLazyServices creates a new LazyServices instance
func NewLazyServices() *LazyServices {
	return &LazyServices{
		AIService:                   NewLazyService(func(ctx context.Context) (*application.AIService, error) { return nil, nil }),
		ContextService:              NewLazyService(func(ctx context.Context) (*contextservice.Service, error) { return nil, nil }),
		ContextBuilder:              NewLazyService(func(ctx context.Context) (domain.ContextBuilder, error) { return nil, nil }),
		ContextAnalyzer:             NewLazyService(func(ctx context.Context) (domain.ContextAnalyzer, error) { return nil, nil }),
		TestService:                 NewLazyService(func(ctx context.Context) (domain.ITestService, error) { return nil, nil }),
		StaticAnalyzerService:       NewLazyService(func(ctx context.Context) (domain.IStaticAnalyzerService, error) { return nil, nil }),
		SBOMService:                 NewLazyService(func(ctx context.Context) (*application.SBOMService, error) { return nil, nil }),
		BuildService:                NewLazyService(func(ctx context.Context) (domain.IBuildService, error) { return nil, nil }),
		VerificationPipelineService: NewLazyService(func(ctx context.Context) (*application.VerificationPipelineService, error) { return nil, nil }),
		ExportService:               NewLazyService(func(ctx context.Context) (*application.ExportService, error) { return nil, nil }),
		ReportService:               NewLazyService(func(ctx context.Context) (*application.ReportService, error) { return nil, nil }),
		TaskProtocolService:         NewLazyService(func(ctx context.Context) (domain.TaskProtocolService, error) { return nil, nil }),
		ErrorAnalyzer:               NewLazyService(func(ctx context.Context) (domain.ErrorAnalyzer, error) { return nil, nil }),
		CorrectionEngine:            NewLazyService(func(ctx context.Context) (domain.CorrectionEngine, error) { return nil, nil }),
		SymbolGraph:                 NewLazyService(func(ctx context.Context) (*application.SymbolGraphService, error) { return nil, nil }),
		RepairService:               NewLazyService(func(ctx context.Context) (domain.RepairService, error) { return nil, nil }),
		GuardrailService:            NewLazyService(func(ctx context.Context) (domain.GuardrailService, error) { return nil, nil }),
		TaskflowService:             NewLazyService(func(ctx context.Context) (domain.TaskflowService, error) { return nil, nil }),
		UXMetricsService:            NewLazyService(func(ctx context.Context) (domain.UXMetricsService, error) { return nil, nil }),
		ApplyService:                NewLazyService(func(ctx context.Context) (*application.ApplyService, error) { return nil, nil }),
		DiffService:                 NewLazyService(func(ctx context.Context) (*application.DiffService, error) { return nil, nil }),
		RouterLLMService:            NewLazyService(func(ctx context.Context) (*application.RouterLLMService, error) { return nil, nil }),
	}
}
