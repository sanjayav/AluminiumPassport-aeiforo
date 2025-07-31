
package services

// CalculateESGScore evaluates an ESG score from emissions and recycled content
func CalculateESGScore(carbonEmissions float64, recycledContent float64) float64 {
    // Normalize inputs
    emissionScore := 100.0 - carbonEmissions*10.0   // Lower is better
    recycleScore := recycledContent * 1.0           // Higher is better

    // Clamp values to [0, 100]
    if emissionScore < 0 {
        emissionScore = 0
    } else if emissionScore > 100 {
        emissionScore = 100
    }

    if recycleScore > 100 {
        recycleScore = 100
    }

    // Weighted average
    return 0.6*emissionScore + 0.4*recycleScore
}
