package models

type VaultRecord struct {
	Id      string
	Key     string
	Hashsum string
}

type VaultFileMeta struct {
	FileId      string
	Filename    string
	ContentType string
	Key         string
	Hashsum     string
}

type VaultFileBase struct {
	MessageId string
}

type AppendFilesArgs struct {
	VaultFileBase
	Files []VaultFileMeta
}

type VaultFile struct {
	VaultFileBase
	VaultFileMeta
}
