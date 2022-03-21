package main

import (
	"github.com/agility323/liberty/demo_service/avatar/avatardata"
)

var AvatarRpcList = []string {"CMD_interact_finish", "CMD_interact_get_reward", }

func (a *Avatar) CMD_interact_finish(id string) {
	if st, _ := a.data.Interacts[id]; st != avatardata.InteractStateGet {
		return
	}
	a.data.Interacts[id] = avatardata.InteractStateFinished
	a.updateProp("interacts", a.data.Interacts)
}

func (a *Avatar) CMD_interact_get_reward(id string) {
	if st, _ := a.data.Interacts[id]; st != avatardata.InteractStateFinished {
		return
	}
	a.data.Interacts[id] = avatardata.InteractStateRewarded
	if jumps, _ := avatardata.InteractJumpMap[id]; jumps != nil {
		for _, j := range jumps {
			a.data.Interacts[j] = avatardata.InteractStateGet
		}
	}
	a.updateProp("interacts", a.data.Interacts)
	a.showReward(avatardata.InteractRewardMap[id])
}
