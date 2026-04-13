package service

import "github.com/harshpn/taskflow/internal/store"

func userFromStore(user store.User) User {
	return User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func usersFromStore(users []store.User) []User {
	result := make([]User, 0, len(users))
	for _, user := range users {
		result = append(result, userFromStore(user))
	}
	return result
}

func projectFromStore(project store.Project) Project {
	return Project{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
	}
}

func projectsFromStore(projects []store.Project) []Project {
	result := make([]Project, 0, len(projects))
	for _, project := range projects {
		result = append(result, projectFromStore(project))
	}
	return result
}

func taskFromStore(task store.Task) Task {
	return Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		ProjectID:   task.ProjectID,
		AssigneeID:  task.AssigneeID,
		CreatorID:   task.CreatorID,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func tasksFromStore(tasks []store.Task) []Task {
	result := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, taskFromStore(task))
	}
	return result
}

func projectStatsFromStore(s store.ProjectStats) ProjectStats {
	counts := make([]AssigneeCount, 0, len(s.AssigneeCounts))
	for _, ac := range s.AssigneeCounts {
		counts = append(counts, AssigneeCount{UserID: ac.UserID, Name: ac.Name, Count: ac.Count})
	}
	return ProjectStats{
		StatusCounts:   s.StatusCounts,
		AssigneeCounts: counts,
	}
}
