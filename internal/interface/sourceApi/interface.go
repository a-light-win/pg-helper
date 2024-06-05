package sourceApi

type SourceAdder interface {
	AddDatabaseSource(source *DatabaseSource) error
}

type SourceRemover interface {
	MarkDatabaseSourceIdle(name string) error
}

type SourceGetter interface {
	IsReady(name string, instName string) bool
	GetSource(name string) *DatabaseSource
}

type SourceHandler interface {
	SourceAdder
	SourceRemover
	SourceGetter
}
