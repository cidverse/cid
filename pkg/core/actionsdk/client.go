package actionsdk

// SDKClient defines the interface on how actions can interact with CID.
type SDKClient interface {
	// misc operations

	HealthV1() (HealthV1Response, error) // simple health check to verify connectivity and basic functionality of the SDK
	LogV1(req LogV1Request) error        // log a message with the given level and context (context is optional key-value pairs that can be included in the log)
	UUIDV4() string                      // UUID generates a new UUID string (useful for generating unique IDs for things like temporary files, correlation IDs, etc.)

	// config operations

	ConfigV1() (*ConfigV1Response, error)
	ProjectExecutionContextV1() (*ProjectExecutionContextV1Response, error)
	ModuleExecutionContextV1() (*ModuleExecutionContextV1Response, error)
	EnvironmentV1() (*EnvironmentV1Response, error)
	DeploymentV1() (*DeploymentV1Response, error)
	ModuleListV1() ([]*ProjectModule, error)
	ModuleCurrentV1() (*ProjectModule, error)

	// Command operations

	ExecuteCommandV1(req ExecuteCommandV1Request) (*ExecuteCommandV1Response, error)

	// VCS operations

	VCSCommitsV1(request VCSCommitsRequest) ([]*VCSCommit, error)
	VCSCommitByHashV1(request VCSCommitByHashRequest) (*VCSCommit, error)
	VCSTagsV1() ([]VCSTag, error)
	VCSReleasesV1(request VCSReleasesRequest) ([]VCSRelease, error)
	VCSDiffV1(request VCSDiffRequest) ([]VCSDiff, error) // VCSDiff generates a diff between two VCS references (like commits or tags) and returns a list of changed files with their change types (added, modified, deleted).

	// File operations

	FileReadV1(file string) (string, error)                 // FileRead reads the content of the specified file and returns it as a string. It returns an error if the operation fails.
	FileWriteV1(file string, content []byte) error          // FileWrite writes the given content to the specified file. It returns an error if the operation fails.
	FileRemoveV1(file string) error                         // FileRemove removes the specified file. It returns an error if the operation fails.
	FileCopyV1(old string, new string) error                // FileCopy copies a file from old path to new path. It returns an error if the operation fails.
	FileRenameV1(old string, new string) error              // FileRename renames a file from old path to new path. It returns an error if the operation fails.
	FileListV1(req FileV1Request) (files []File, err error) // FileList lists files in a directory based on the given request parameters (like directory path and optional extensions filter). It returns a slice of files and an error if the operation fails.
	FileExistsV1(file string) bool                          // FileExists checks if the specified file exists and returns true if it does, false otherwise.

	// Artifact operations

	ArtifactListV1(request ArtifactListRequest) ([]*Artifact, error)                                                // ArtifactList lists artifacts based on the given query expression (which can filter artifacts by their metadata). It returns a slice of artifacts that match the query and an error if the operation fails.
	ArtifactByIdV1(id string) (*Artifact, error)                                                                    // ArtifactById retrieves the artifact with the specified ID. It returns the artifact and an error if the operation fails or if the artifact is not found.
	ArtifactUploadV1(request ArtifactUploadRequest) (filePath string, fileHash string, err error)                   // ArtifactUpload uploads a file as an artifact with the specified metadata (like module, type, format, etc.). It returns the path of the stored artifact, its hash, and an error if the operation fails.
	ArtifactDownloadV1(request ArtifactDownloadRequest) (*ArtifactDownloadResult, error)                            // ArtifactDownload downloads the artifact with the specified ID to the target file path. It returns the path of the downloaded file, its hash, and size, or an error if the operation fails.
	ArtifactDownloadByteArrayV1(request ArtifactDownloadByteArrayRequest) (*ArtifactDownloadByteArrayResult, error) // ArtifactDownloadByteArray downloads the artifact with the specified ID and returns its content as a byte array along with its hash and size. It returns an error if the operation fails.

	// archive operations

	ZIPCreateV1(inputDirectory string, outputFile string) error    // ZIPCreate creates a zip archive of the directory at the given path. It takes the input directory and the output file path for the zip archive. It returns an error if the operation fails.
	ZIPExtractV1(archiveFile string, outputDirectory string) error // ZIPExtract unzips the zip archive at the given path into the given directory. It takes the path of the zip archive and the target output directory. It returns an error if the operation fails.
	TARCreateV1(inputDirectory string, outputFile string) error    // TARCreate creates a tar archive of the directory at the given path. It takes the input directory and the output file path for the tar archive. It returns an error if the operation fails.
	TARExtractV1(archiveFile string, outputDirectory string) error // TARExtract extracts a tar archive at the given path into the given directory. It takes the path of the tar archive and the target output directory. It returns an error if the operation fails.
}
