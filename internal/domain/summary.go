package domain

func StatusCounts(deployment_jobs []DeploymentJob) map[DeploymentJobStatus]int {
	counts := make(map[DeploymentJobStatus]int)
	for _, task := range deployment_jobs {
		counts[task.Status]++
	}
	return counts
}
