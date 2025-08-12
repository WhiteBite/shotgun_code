package application

import (
	"context"
	"shotgun_code/domain"
)

type ProjectService struct {
	log                      domain.Logger
	bus                      domain.EventBus
	settingsService          *SettingsService
	gitRepo                  domain.GitRepository
	treeBuilder              domain.FileTreeBuilder
	diffSplitter             domain.DiffSplitter
	contextGenerationService *ContextGenerationService
	isGitAvailable           bool
}

func NewProjectService(
	log domain.Logger,
	bus domain.EventBus,
	settingsService *SettingsService,
	gitRepo domain.GitRepository,
	treeBuilder domain.FileTreeBuilder,
	diffSplitter domain.DiffSplitter,
	contextGenService *ContextGenerationService,
) (*ProjectService, error) {
	s := &ProjectService{
		log:                      log,
		bus:                      bus,
		settingsService:          settingsService,
		gitRepo:                  gitRepo,
		treeBuilder:              treeBuilder,
		diffSplitter:             diffSplitter,
		contextGenerationService: contextGenService,
	}
	var err error
	s.isGitAvailable, err = s.gitRepo.CheckAvailability()
	if err != nil {
		s.log.Error("Error checking Git availability: " + err.Error())
	}
	if s.isGitAvailable {
		s.log.Info("Git is available on the system.")
	} else {
		s.log.Warning("Git not found. Git-related functionality will be disabled.")
	}
	return s, nil
}

func (s *ProjectService) LogError(message string) {
	s.log.Error(message)
	s.bus.Emit("app:error", message)
}

func (s *ProjectService) ListFiles(dirPath string, useGitignore bool, useCustomIgnore bool) ([]*domain.FileNode, error) {
	return s.treeBuilder.BuildTree(dirPath, useGitignore, useCustomIgnore)
}

func (s *ProjectService) IsGitAvailable() bool {
	return s.isGitAvailable
}

func (s *ProjectService) GetUncommittedFiles(projectRoot string) ([]domain.FileStatus, error) {
	return s.gitRepo.GetUncommittedFiles(projectRoot)
}

func (s *ProjectService) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	return s.gitRepo.GetRichCommitHistory(projectRoot, branchName, limit)
}

func (s *ProjectService) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	s.contextGenerationService.Generate(ctx, rootDir, includedPaths)
}

func (s *ProjectService) SplitShotgunDiff(gitDiffText string, approxLineLimit int) ([]string, error) {
	return s.diffSplitter.Split(gitDiffText, approxLineLimit)
}
