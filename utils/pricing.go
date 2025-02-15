package utils

// CalculateFare calculates the fare based on distance in kilometers
func CalculateFare(distance float64) float64 {
	if distance < 0 {
		return 0
	}

	// Fixed price brackets
	if distance <= 2.5 {
		return 10.0
	}
	if distance <= 5.0 {
		return 15.0
	}
	if distance <= 7.5 {
		return 20.0
	}
	if distance <= 10.0 {
		return 25.0
	}
	if distance <= 12.5 {
		return 30.0
	}
	if distance <= 15.0 {
		return 35.0
	}
	if distance <= 17.5 {
		return 40.0
	}
	if distance <= 20.0 {
		return 45.0
	}
	if distance <= 22.5 {
		return 50.0
	}
	if distance <= 25.0 {
		return 55.0
	}
	if distance <= 27.5 {
		return 60.0
	}
	if distance <= 30.0 {
		return 65.0
	}

	// For distances over 30km, use rate per kilometer
	return float64(int(distance * 2.17)) // Round to nearest integer
}
