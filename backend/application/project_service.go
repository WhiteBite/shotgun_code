package application

import (
	"context"
	"shotgun_code/domain"
)

type ProjectService struct {
	log                      domain.Logger
	bus                      domain.EventBus
	treeBuilder              domain.TreeBuilder
	gitRepo                  domain.GitRepository
	contextGenerationService *ContextGenerationService
}

func NewProjectService(
	log domain.Logger,
	bus domain.EventBus,
	treeBuilder domain.TreeBuilder,
	gitRepo domain.GitRepository,
	contextGenerationService *ContextGenerationService,
) *ProjectService {
	return &ProjectService{
		log:                      log,
		bus:                      bus,
		treeBuilder:              treeBuilder,
		gitRepo:                  gitRepo,
		contextGenerationService: contextGenerationService,
	}
}

func (s *ProjectService) ListFiles(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	s.log.Info("Listing files for directory: " + dirPath)
	nodes, err := s.treeBuilder.BuildTree(dirPath, useGitignore, useCustomIgnore)
	if err != nil {
		s.log.Error("Failed to build file tree: " + err.Error())
		return nil, err
	}
	s.log.Info("File tree built successfully")
	return nodes, nil
}

func (s *ProjectService) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	s.log.Info("Getting uncommitted files for: " + projectRoot)
	files, err := s.gitRepo.GetUncommittedFiles(projectRoot)
	if err != nil {
		s.log.Error("Failed to get uncommitted files: " + err.Error())
		return nil, err
	}
	return files, nil
}

func (s *ProjectService) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	s.log.Info("Getting commit history for: " + projectRoot)
	commits, err := s.gitRepo.GetRichCommitHistory(projectRoot, branchName, limit)
	if err != nil {
		s.log.Error("Failed to get commit history: " + err.Error())
		return nil, err
	}
	return commits, nil
}

func (s *ProjectService) IsGitAvailable() bool {
	return s.gitRepo.IsGitAvailable()
}

func (s *ProjectService) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	s.log.Info("Starting context generation")
	// Теперь используем безопасный метод с panic recovery
	s.contextGenerationService.GenerateContext(ctx, rootDir, includedPaths)
}
