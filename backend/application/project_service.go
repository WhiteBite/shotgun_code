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
		s.log.Error("Произошла ошибка при проверке доступности Git: " + err.Error())
	}
	if s.isGitAvailable {
		s.log.Info("Git доступен в системе.")
	} else {
		s.log.Warning("Git не найден. Функционал, связанный с Git, будет отключен.")
	}
	return s, nil
}

func (s *ProjectService) LogError(message string) {
	s.log.Error(message)
	s.bus.Emit("app:error", message)
}

func (s *ProjectService) ListFiles(dirPath string) ([]*domain.FileNode, error) {
	useGit := s.settingsService.GetUseGitignore()
	useCustom := s.settingsService.GetUseCustomIgnore()
	return s.treeBuilder.BuildTree(dirPath, useGit, useCustom)
}

func (s *ProjectService) IsGitAvailable() bool {
	return s.isGitAvailable
}

func (s *ProjectService) GetUncommittedFiles(projectRoot string) ([]string, error) {
	return s.gitRepo.GetUncommittedFiles(projectRoot)
}

func (s *ProjectService) GetRichCommitHistory(projectRoot, branchName string, limit int) ([]domain.CommitWithFiles, error) {
	return s.gitRepo.GetRichCommitHistory(projectRoot, branchName, limit)
}

// GenerateContext delegates the context generation to the specialized service.
func (s *ProjectService) GenerateContext(ctx context.Context, rootDir string, includedPaths []string) {
	s.contextGenerationService.Generate(ctx, rootDir, includedPaths)
}

func (s *ProjectService) SplitShotgunDiff(gitDiffText string, approxLineLimit int) ([]string, error) {
	return s.diffSplitter.Split(gitDiffText, approxLineLimit)
}
