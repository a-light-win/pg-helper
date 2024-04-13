package job

type JobProducer struct {
	AddJobs chan Job
}

func (j *JobProducer) Produce(job Job) {
	j.AddJobs <- job
}
