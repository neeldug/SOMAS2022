package team3

import (
	"infra/game/agent"
	"infra/game/commons"
	"infra/game/decision"
	"infra/game/message"
	"infra/game/message/proposal"
	"math/rand"
	"sort"

	"github.com/benbjohnson/immutable"
	"golang.org/x/exp/maps"
)

func (a *AgentThree) LootActionNoProposal(baseAgent agent.BaseAgent) immutable.SortedMap[commons.ItemID, struct{}] {
	loot := baseAgent.Loot()
	weapons := loot.Weapons().Iterator()
	shields := loot.Shields().Iterator()
	hpPotions := loot.HpPotions().Iterator()
	staminaPotions := loot.StaminaPotions().Iterator()

	builder := immutable.NewSortedMapBuilder[commons.ItemID, struct{}](nil)

	for !weapons.Done() {
		weapon, _ := weapons.Next()
		if rand.Int()%2 == 0 {
			builder.Set(weapon.Id(), struct{}{})
		}
	}

	for !shields.Done() {
		shield, _ := shields.Next()
		if rand.Int()%2 == 0 {
			builder.Set(shield.Id(), struct{}{})
		}
	}

	for !hpPotions.Done() {
		pot, _ := hpPotions.Next()
		if rand.Int()%2 == 0 {
			builder.Set(pot.Id(), struct{}{})
		}
	}

	for !staminaPotions.Done() {
		pot, _ := staminaPotions.Next()
		if rand.Int()%2 == 0 {
			builder.Set(pot.Id(), struct{}{})
		}
	}

	return *builder.Map()
}

// function to take loot when given access to pool
func (a *AgentThree) LootAction(
	baseAgent agent.BaseAgent,
	proposedLoot immutable.SortedMap[commons.ItemID, struct{}],
	acceptedProposal message.Proposal[decision.LootAction],
) immutable.SortedMap[commons.ItemID, struct{}] {
	// take allocated loot
	return proposedLoot
}

// Never called
func (a *AgentThree) HandleLootInformation(m message.TaggedInformMessage[message.LootInform], baseAgent agent.BaseAgent) {
	// submit a proposal to the leader
	switch m.Message().(type) {
	case message.LootInform:
		// Send Proposal?
		sendProposal := rand.Intn(100)
		if sendProposal < a.personality {
			// general and send a loot proposal
			baseAgent.SendLootProposalToLeader(a.generateLootProposal(baseAgent))
		}
	default:
		return
	}
}

// forcibly call at start of loot phase to begin proceedings
func (a *AgentThree) RequestLootProposal(baseAgent agent.BaseAgent) { // put your logic here now, instead
	sendProposal := rand.Intn(100)
	if sendProposal > a.personality {
		return
	}
	// generate and send a loot proposal at the start of every turn
	baseAgent.SendLootProposalToLeader(a.generateLootProposal(baseAgent))
}

func (a *AgentThree) HandleLootProposal(_ message.Proposal[decision.LootAction], _ agent.BaseAgent) decision.Intent {
	// vote on the loot proposal
	// do i vote?
	toVote := rand.Intn(100)
	if toVote < a.personality {
		// Enter logic for evaluating a loot proposal here
		switch rand.Intn(2) {
		case 0:
			return decision.Positive
		default:
			return decision.Negative
		}
	} else {
		// abstain
		return decision.Abstain
	}
}

func (a *AgentThree) generateLootProposal(baseAgent agent.BaseAgent) commons.ImmutableList[proposal.Rule[decision.LootAction]] {
	rules := make([]proposal.Rule[decision.LootAction], 0)
	HpPotionThresh, StPotionThresh, SwordThresh, ShieldThresh := a.LootThresholdDecision(baseAgent)

	// health potion rule
	rules = append(rules, *proposal.NewRule(decision.HealthPotion,
		proposal.NewComparativeCondition(proposal.Health, proposal.LessThan, uint(HpPotionThresh))))
	// weapon rules
	rules = append(rules, *proposal.NewRule(decision.Weapon,
		proposal.NewAndCondition(*proposal.NewComparativeCondition(proposal.Health, proposal.LessThan, uint(HpPotionThresh)),
			*proposal.NewComparativeCondition(proposal.TotalAttack, proposal.LessThan, uint(SwordThresh)),
		)))
	// stamina potion rule
	rules = append(rules, *proposal.NewRule(decision.StaminaPotion,
		proposal.NewComparativeCondition(proposal.Stamina, proposal.LessThan, uint(StPotionThresh))))
	// shield rules
	rules = append(rules, *proposal.NewRule(decision.Shield,
		proposal.NewAndCondition(*proposal.NewComparativeCondition(proposal.Health, proposal.LessThan, uint(HpPotionThresh)),
			*proposal.NewComparativeCondition(proposal.TotalAttack, proposal.LessThan, uint(ShieldThresh)),
		)))

	return *commons.NewImmutableList(rules)
}

// determin loot thresholds
func (a *AgentThree) LootThresholdDecision(baseAgent agent.BaseAgent) (float64, float64, float64, float64) {
	HPThreshold, StaminaThreshold, AttackThreshold, DefenseThreshold := 0.0, 0.0, 0.0, 0.0
	// initiate modifers
	alpha := 0.2
	beta := 0.1
	// extract agents
	agentState := baseAgent.AgentState()
	// get my stats
	myHP := float64(agentState.Hp)
	myST := float64(agentState.Stamina)
	myATT := float64(agentState.Attack)
	myDEF := float64(agentState.Defense)
	// get group stats
	groupAvHP := AverageArray(GetHealthAllAgents(baseAgent))
	groupAvST := AverageArray(GetStaminaAllAgents(baseAgent))
	groupAvATT := AverageArray(GetAttackAllAgents(baseAgent))
	groupAvDEF := AverageArray(GetDefenceAllAgents(baseAgent))

	// get differences (group to me)
	Delta1HP := groupAvHP - float64(myHP)
	Delta1ST := groupAvST - float64(myST)
	Delta1ATT := groupAvDEF - float64(myATT)
	Delta1DEF := groupAvATT - float64(myDEF)

	if len(a.TSN) > 0 {
		// get TSN average stats
		TSNavHP := AverageArray(GetHealthTSN(baseAgent, a.TSN))
		TSNavST := AverageArray(GetStaminaTSN(baseAgent, a.TSN))
		TSNavATT := AverageArray(GetAttackTSN(baseAgent, a.TSN))
		TSNavDEF := AverageArray(GetDefenceTSN(baseAgent, a.TSN))
		// get differences (group to TSN)
		Delta2HP := groupAvHP - TSNavHP
		Delta2ST := groupAvST - TSNavST
		Delta2ATT := groupAvATT - TSNavATT
		Delta2DEF := groupAvDEF - TSNavDEF

		HPThreshold = myHP + alpha*Delta1HP + beta*Delta2HP
		StaminaThreshold = myST + alpha*Delta1ST + beta*Delta2ST
		AttackThreshold = myATT + alpha*Delta1ATT + beta*Delta2ATT
		DefenseThreshold = myDEF + alpha*Delta1DEF + beta*Delta2DEF

		return HPThreshold, StaminaThreshold, AttackThreshold, DefenseThreshold
	}
	// caluclate the thresholds (for all the decisions)
	HPThreshold = (myHP + alpha*Delta1HP) * float64(1.02)
	StaminaThreshold = (myST + alpha*Delta1ST) * float64(1.02)
	AttackThreshold = (myATT + alpha*Delta1ATT) * float64(1.05)
	DefenseThreshold = (myDEF + alpha*Delta1DEF) * float64(1.05)

	return HPThreshold, StaminaThreshold, AttackThreshold, DefenseThreshold
}

// func (a *AgentThree) ChooseItem(baseAgent agent.BaseAgent,
// 	items map[string]struct{}, weaponSet map[string]uint, shieldSet map[string]uint, hpPotionSet map[string]uint, staminaPotionSet map[string]uint) string {
// 	// function to calculate the agents choice of loot

// 	// get group average stats
// 	avHP, avST, avATT, avDEF := GetGroupAv(baseAgent)
// 	// normalise the group stats
// 	groupAvHP, groupAvST, groupAvATT, groupAvDEF := normalize4El(avHP, avST, avATT, avDEF)
// 	// get agent
// 	agentState := baseAgent.AgentState()
// 	HP := float64(agentState.Hp)
// 	ST := float64(agentState.Stamina)
// 	ATT := float64(agentState.BonusAttack())
// 	DEF := float64(agentState.BonusDefense())
// 	// normalise the agent stats
// 	meanHP, meanST, meanATT, meanDEF := normalize4El(HP, ST, ATT, DEF)

// 	// cal differences
// 	diffHP := groupAvHP - meanHP
// 	diffST := groupAvST - meanST
// 	diffATT := groupAvATT - meanATT
// 	diffDEF := groupAvDEF - meanDEF

// 	// create an array of the above, order them
// 	diffs := []float64{diffHP, diffST, diffATT, diffDEF}
// 	sortedDiffs := make([]float64, len(diffs))
// 	copy(sortedDiffs, diffs)
// 	sort.Slice(sortedDiffs, func(i, j int) bool {
// 		return sortedDiffs[i] > sortedDiffs[j]
// 	})
// 	var item string
// 	// return the item that is needed most (out of the items available)
// 	for _, val := range sortedDiffs {
// 		if val == 0 {
// 			// if val is zero, everyone the same so take arbitrary loot
// 			for id := range items {
// 				item = id
// 				break
// 			}
// 			return item
// 		} else if val == diffHP {
// 			//search of item in corresponding set
// 			item = searchForItem(hpPotionSet, items)
// 			// if excists, then return the item
// 			if len(item) > 0 {
// 				return item
// 			}
// 		} else if val == diffST {
// 			item = searchForItem(staminaPotionSet, items)
// 			if len(item) > 0 {
// 				return item
// 			}
// 		} else if val == diffATT {
// 			item = searchForItem(weaponSet, items)
// 			if len(item) > 0 {
// 				return item
// 			}
// 		} else if val == diffDEF {
// 			item = searchForItem(shieldSet, items)
// 			if len(item) > 0 {
// 				return item
// 			}
// 		}
// 	}
// 	if item == "" {
// 		// if got nothing, take arbitrary loot
// 		for id := range items {
// 			item = id
// 			break
// 		}
// 	}
// 	return item
// }

func (a *AgentThree) ChooseItem(baseAgent agent.BaseAgent,
	items map[string]struct{}, weaponSet map[string]uint, shieldSet map[string]uint, hpPotionSet map[string]uint, staminaPotionSet map[string]uint) string {
	// function to calculate the agents choice of loot

	// get group average stats
	avHP, avST, avATT, avDEF := GetGroupAv(baseAgent)
	// normalise the group stats
	groupAvHP, groupAvST, groupAvATT, groupAvDEF := normalize4El(avHP, avST, avATT, avDEF)
	// get agent
	agentState := baseAgent.AgentState()
	HP := float64(agentState.Hp)
	ST := float64(agentState.Stamina)
	ATT := float64(agentState.BonusAttack())
	DEF := float64(agentState.BonusDefense())
	// normalise the agent stats
	meanHP, meanST, meanATT, meanDEF := normalize4El(HP, ST, ATT, DEF)

	// cal differences
	diffHP := groupAvHP - meanHP
	diffST := groupAvST - meanST
	diffATT := groupAvATT - meanATT
	diffDEF := groupAvDEF - meanDEF

	// create an array of the above, order them
	diffs := []float64{diffHP, diffST, diffATT, diffDEF}
	sortedDiffs := make([]float64, len(diffs))
	copy(sortedDiffs, diffs)
	sort.Slice(sortedDiffs, func(i, j int) bool {
		return sortedDiffs[i] > sortedDiffs[j]
	})

	defaultItem := maps.Keys(items)[0]
	var activeSet map[string]uint

	for _, val := range sortedDiffs {
		switch val {
		case 0:
			return defaultItem
		case diffHP:
			activeSet = hpPotionSet
		case diffST:
			activeSet = staminaPotionSet
		case diffATT:
			activeSet = weaponSet
		case diffDEF:
			activeSet = shieldSet
		default:
		}
		potentialItem := searchForItem(activeSet, items)

		if potentialItem != "" {
			return potentialItem
		}
	}

	return defaultItem
}

func searchForItem(set map[string]uint, items map[string]struct{}) string {
	for item := range items {
		if _, ok := set[item]; ok {
			return item
		}
	}
	return ""
}

func GetGroupAv(baseAgent agent.BaseAgent) (float64, float64, float64, float64) {
	avHP := AverageArray(GetHealthAllAgents(baseAgent))
	avST := AverageArray(GetStaminaAllAgents(baseAgent))
	avATT := AverageArray(GetAttackAllAgents(baseAgent))
	avDEF := AverageArray(GetDefenceAllAgents(baseAgent))

	return avHP, avST, avATT, avDEF
}
