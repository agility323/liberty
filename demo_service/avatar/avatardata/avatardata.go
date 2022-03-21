package avatardata

type AvatarData struct {
	// persistent
	// basic
	Uid int `bson:"uid" msgpack:"uid"`
	Name string	`bson:"name" msgpack:"name"`
	Account string `bson:"account" msgpack:"account"`
	// develop
	Exp int `bson:"exp" msgpack:"exp"`
	Lv int `bson:"lv" msgpack:"lv"`
	// resource
	Gold int `bson:"gold" msgpack:"gold"`
	// gameplay
	Buildings map[string]*BuildingProp `bson:"buildings,omitempty" msgpack:"buildings,omitempty"`
	Interacts map[string]int `bson:"interacts,omitempty" msgpack:"interacts,omitempty"`

	// non persistent
	OnlineTime int `bson:",-" msgpack:"online_time"`
}


