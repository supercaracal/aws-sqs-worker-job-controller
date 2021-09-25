package worker

import (
	batchv1 "k8s.io/api/batch/v1"
)

// JobsOrderedByStartTimeASC is
type JobsOrderedByStartTimeASC []*batchv1.Job

func (aj JobsOrderedByStartTimeASC) Len() int {
	return len(aj)
}

func (aj JobsOrderedByStartTimeASC) Swap(i, j int) {
	aj[i], aj[j] = aj[j], aj[i]
}

func (aj JobsOrderedByStartTimeASC) Less(i, j int) bool {
	if aj[i].Status.StartTime == nil && aj[j].Status.StartTime != nil {
		return false
	}

	if aj[i].Status.StartTime != nil && aj[j].Status.StartTime == nil {
		return true
	}

	if aj[i].Status.StartTime.Equal(aj[j].Status.StartTime) {
		return aj[i].Name < aj[j].Name
	}

	return aj[i].Status.StartTime.Before(aj[j].Status.StartTime)
}
