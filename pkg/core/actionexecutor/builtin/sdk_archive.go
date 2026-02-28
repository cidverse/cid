package builtin

import "github.com/cidverse/cidverseutils/compress"

// ZIPCreate creates a zip archive of the directory at the given path.
func (sdk ActionSDK) ZIPCreateV1(inputDirectory string, outputFile string) error {
	return compress.ZIPCreate(inputDirectory, outputFile)
}

// ZIPExtract unzips the zip archive at the given path into the given directory.
func (sdk ActionSDK) ZIPExtractV1(archiveFile string, outputDirectory string) error {
	return compress.ZIPExtract(archiveFile, outputDirectory)
}

// TARCreate creates a tar archive of the directory at the given path.
func (sdk ActionSDK) TARCreateV1(inputDirectory string, outputFile string) error {
	return compress.TARCreate(inputDirectory, outputFile)
}

// TARExtract extracts a tar archive at the given path into the given directory.
func (sdk ActionSDK) TARExtractV1(archiveFile string, outputDirectory string) error {
	return compress.TARExtract(archiveFile, outputDirectory)
}
