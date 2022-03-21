package avatardata

const (
	BUILDING_CASTLE = "1001"

	BUILDING_FARM = "2001"
	BUILDING_HOUSE = "2002"
	BUILDING_WATERWHEEL = "2003"
	BUILDING_GATE = "2004"
	BUILDING_RUIN1 = "2005"

)

const (
	BUILDNG_STATE_BITMASK_BUILDING = 1 << iota
	BUILDNG_STATE_BITMASK_PRODUCING
	BUILDNG_STATE_BITMASK_COLLECT
	BUILDNG_STATE_BITMASK_DESTROYED
)

type BuildingProp struct {
	State uint16 `bson:"state" msgpack:"state"`
	Lv uint16 `bson:"lv" msgpack:"lv"`
	X int16 `bson:"x" msgpack:"x"`
	Y int16 `bson:"y" msgpack:"y"`
}
