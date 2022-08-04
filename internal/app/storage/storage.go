package storage

type GotchaStorage interface {
	FindUserByNickname(username string)
}

