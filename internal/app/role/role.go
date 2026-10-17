package role

type Role int

const (
	Astronomer Role = iota // 0 - обычный пользователь / астроном / создатель услуг
	Moderator              // 1 - модератор
)

// Синонимы для совместимости с методичкой и ролевыми моделями
const (
	Creator Role = Astronomer
	Buyer   Role = Astronomer
	Admin   Role = Moderator
	Manager Role = Moderator
)

func (r Role) String() string {
	switch r {
	case Astronomer:
		return "astronomer"
	case Moderator:
		return "moderator"
	default:
		return "unknown"
	}
}
