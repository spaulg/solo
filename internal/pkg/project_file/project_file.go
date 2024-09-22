package project_file

type ProjectFile struct {
	FilePath string
}

func New(projectFilePath string) *ProjectFile {
	return &ProjectFile{
		FilePath: projectFilePath,
	}
}
