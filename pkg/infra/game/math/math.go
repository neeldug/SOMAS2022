package math

import (
	"math"
	"math/rand"
	"time"

	"infra/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Enemy Resilience Modifier
func CalculateDelta() float64 {
	min := 0.8
	max := 1.2
	return min + rand.Float64()*(max-min)
}

// X, the monster’s resilience
func CalculateMonsterHealth(nAgent uint, stamina uint, nLevel uint, currentLevel uint) uint {
	delta := CalculateDelta()
	NFp := float64(nAgent)
	LFp := float64(nLevel)

	agentStaminaRatio := NFp / LFp * float64(stamina)
	levelRatio := float64(currentLevel)/LFp + 0.5

	healthBoost := 1.3

	totalHealth := delta * agentStaminaRatio * levelRatio

	return uint(totalHealth * healthBoost)

	// return uint(math.Ceil((float64(nAgent) * float64(stamina) / float64(nLevel)) * delta * (float64(currentLevel)/float64(nLevel) + 0.5)))
}

// Y, monster’s damage rating
func CalculateMonsterDamage(nAgent uint, HP uint, stamina uint, thresholdPercentage float32, nLevel uint, currentLevel uint) uint {
	delta := CalculateDelta()
	NFp := float64(nAgent)
	LFp := float64(nLevel)
	damageBoost := 2.5

	agentRatio := NFp / LFp
	hpStamSum := float64(HP) + float64(stamina)
	levelRatio := float64(currentLevel)/LFp + 0.5

	totalDamage := delta * agentRatio * hpStamSum * levelRatio

	return uint(totalDamage * damageBoost)

	// return uint(delta * (NFp / LFp) * (float64(HP) + float64(stamina)) * (float64(currentLevel)/LFp + 0.5))
}

func GetNextLevelMonsterValues(gameConfig config.GameConfig, currentLevel uint) (uint, uint) {
	return CalculateMonsterHealth(gameConfig.InitialNumAgents, gameConfig.Stamina, gameConfig.NumLevels, currentLevel+1), CalculateMonsterDamage(gameConfig.InitialNumAgents, gameConfig.StartingHealthPoints, gameConfig.Stamina, gameConfig.ThresholdPercentage, gameConfig.NumLevels, currentLevel+1)
}

func NumberPotionDropped(P float64, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(delta * P * float64(nAgent))
}

func NumberEquipmentDropped(E float64, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(delta * E * float64(nAgent))
}

// function encapsulated to use same random val
func GetPotionDistribution(nAgent uint) (uint, uint) {
	tau := rand.Float64()
	// hardcoded by design
	P := float64(config.EnvToFloat("POTION_SCARCITY_PCT", 0.2))
	NumberHealthPotionDropped := uint((tau) * float64(NumberPotionDropped(P, nAgent)))
	NumberStaminaPotionDropped := uint((1 - tau) * float64(NumberPotionDropped(P, nAgent)))
	return NumberHealthPotionDropped, NumberStaminaPotionDropped
}

// tau recalculated for equipment  and potions
func GetEquipmentDistribution(nAgent uint) (uint, uint) {
	tau := rand.Float64()
	// hardcoded by design
	E := float64(config.EnvToFloat("EQUIPMENT_SCARCITY_PCT", 0.15))
	NumberWeaponDropped := uint((tau) * float64(NumberEquipmentDropped(E, nAgent)))
	NumberShieldDropped := uint((1 - tau) * float64(NumberEquipmentDropped(E, nAgent)))
	return NumberWeaponDropped, NumberShieldDropped
}

func GetWeaponDamage(X uint, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(math.Ceil((delta * float64(X) * 2) / float64(nAgent)))
}

func GetShieldProtection(Y uint, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(math.Ceil((delta * float64(Y)) / (float64(nAgent))))
}

func GetHealthPotionValue(Y uint, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(math.Ceil((delta * float64(Y) * 5.0) / (float64(nAgent))))
}

func GetStaminaPotionValue(X uint, nAgent uint) uint {
	delta := CalculateDelta()
	return uint(math.Ceil((delta * float64(X) * 5.0) / (float64(nAgent))))
}
