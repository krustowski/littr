package stats

import (
	"context"
	//"fmt"
	"time"

	"go.vxn.dev/littr/pkg/models"
)

type StatService struct {
	pollRepository models.PollRepositoryInterface
	postRepository models.PostRepositoryInterface
	userRepository models.UserRepositoryInterface
}

func NewStatService(
	pollRepository models.PollRepositoryInterface,
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.StatServiceInterface {

	if pollRepository == nil || postRepository == nil || userRepository == nil {
		return nil
	}

	return &StatService{
		pollRepository: pollRepository,
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *StatService) Calculate(ctx context.Context) (*map[string]int64, *map[string]models.UserStat, *map[string]models.User, error) {
	flowStats := make(map[string]int64)
	userStats := make(map[string]models.UserStat)

	polls, err := s.pollRepository.GetAll()
	if err != nil {
		return nil, nil, nil, err
	}

	posts, err := s.postRepository.GetAll()
	if err != nil {
		return nil, nil, nil, err
	}

	users, err := s.userRepository.GetAll()
	if err != nil {
		return nil, nil, nil, err
	}

	flowers := make(map[string]int64)
	shades := make(map[string]int64)

	// Init the flowStats.
	flowStats["posts"] = int64(len(*posts))
	//flowStats["users"] = int64(len(*users))
	flowStats["users"] = -1
	flowStats["stars"] = 0

	// Iterate over all posts, compose stats results.
	for _, val := range *posts {
		// increment user's stats
		stat, ok := userStats[val.Nickname]
		if !ok {
			// Create a blank stat.
			stat = models.UserStat{}
			stat.Searched = true
		}

		// Increase the post count, increase the reaction count sum.
		stat.PostCount++
		stat.ReactionCount += val.ReactionCount

		userStats[val.Nickname] = stat
		flowStats["stars"] += val.ReactionCount
	}

	// Iterate over all users, compose global flower and shade count.
	for _, user := range *users {
		// flower count
		for key, enabled := range user.FlowList {
			if enabled && key != user.Nickname {
				flowers[key]++
			}
		}

		// Calculate the shade count.
		for key, shaded := range user.ShadeList {
			if shaded && key != user.Nickname {
				shades[key]++
			}
		}

		// Check the online status.
		diff := time.Since(user.LastActiveTime)
		if diff < 15*time.Minute {
			flowStats["online"]++
		}

		flowStats["users"]++
	}

	// Iterate over composed flowers, assign the count to an user.
	for key, count := range flowers {
		stat := userStats[key]

		// FlowList also contains the user itself
		stat.FlowerCount = count
		userStats[key] = stat
	}

	// Iterate over compose shades, assign the count to an user.
	for key, count := range shades {
		stat := userStats[key]

		// FlowList also contains the user itself
		stat.ShadeCount = count
		userStats[key] = stat
	}

	// Iterate over all polls, count them good.
	for _, poll := range *polls {
		flowStats["polls"]++

		flowStats["votes"] += poll.OptionOne.Counter
		flowStats["votes"] += poll.OptionTwo.Counter
		flowStats["votes"] += poll.OptionThree.Counter
	}

	return &flowStats, &userStats, users, nil
}
