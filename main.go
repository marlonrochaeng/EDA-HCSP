package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

func get_n_random_elements(length int, min int, max int, n int) [][]int {
	elements := make([][]int, n)
	for i := 0; i < n; i++ {
		// Create an array to hold the random numbers
		randomNumbers := make([]int, length)

		// Fill the array with random numbers
		for j := range randomNumbers {
			randomNumbers[j] = rand.Intn(max-min+1) + min
		}

		elements[i] = randomNumbers
	}

	return elements
}

func get_files_in_path(path string) {
	// ex: "512x16/*"
	files, _ := filepath.Glob(path)

	fmt.Println(files)
}

func generateCombinations(keys []string, values map[string][]int, index int, currentCombination map[string]int, allCombinations *[]map[string]int) {
	if index == len(keys) {
		// Copy the current combination to avoid referencing the same map
		combination := make(map[string]int)
		for k, v := range currentCombination {
			combination[k] = v
		}
		*allCombinations = append(*allCombinations, combination)
		return
	}

	key := keys[index]
	for _, value := range values[key] {
		currentCombination[key] = value
		generateCombinations(keys, values, index+1, currentCombination, allCombinations)
	}
}

func get_ET_as_array(file_name string) []float64 {
	file, _ := os.Open(file_name)
	//file, _ := os.Open("data.txt")

	defer file.Close()

	var floatValues []float64

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Read each line as a string
		line := scanner.Text()

		// Convert the string to a float64
		value, _ := strconv.ParseFloat(line, 64)

		// Append the float64 value to the slice
		floatValues = append(floatValues, value)
	}

	return floatValues
}

func generate_et_and_ct(slice []float64, n, m int) [][]float64 {

	// Create the matrix
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, m)
	}

	// Fill the matrix with values from the slice
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			matrix[i][j] = slice[i*m+j]
		}
	}

	return matrix
}

func printMatrix(matrix [][]float64) {
	for _, row := range matrix {
		for _, val := range row {
			fmt.Printf("%6.2f ", val) // Adjust the format as needed
		}
		fmt.Println()
	}
}

// Function to get the index of the maximum value in a slice
func getMaxInArray(arr []float64) int {
	maxIndex := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] > arr[maxIndex] {
			maxIndex = i
		}
	}
	return maxIndex
}

// Function to calculate fitness
func getFitness(ET [][]float64, individual []int) float64 {
	numMachines := len(ET[0])
	maquinas := make([]float64, numMachines)

	for i := 0; i < len(ET); i++ {
		maquinas[individual[i]] += ET[i][individual[i]]
	}

	return maquinas[getMaxInArray(maquinas)]
}

func orderPop(ET [][]float64, individuals [][]int) [][]int {
	type result struct {
		index   int
		fitness float64
	}

	results := make(chan result, len(individuals))
	var wg sync.WaitGroup

	// Use goroutines to compute fitness in parallel
	for i, individual := range individuals {
		wg.Add(1)
		go func(i int, individual []int) {
			defer wg.Done()
			fitness := getFitness(ET, individual)
			results <- result{index: i, fitness: fitness}
		}(i, individual)
	}

	// Close the results channel once all goroutines have finished
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect the results
	fitnesses := make([]result, 0, len(individuals))
	for res := range results {
		fitnesses = append(fitnesses, res)
	}

	// Sort the results based on fitness
	sort.Slice(fitnesses, func(i, j int) bool {
		return fitnesses[i].fitness < fitnesses[j].fitness
	})

	// Order the individuals based on the sorted fitnesses
	orderedIndividuals := make([][]int, len(individuals))
	for i, res := range fitnesses {
		orderedIndividuals[i] = individuals[res.index]
	}

	return orderedIndividuals
}

func createProbMatrix(jobs, machines int) [][]int {
	// Initialize a 2D slice with zeros
	matrix := make([][]int, jobs)
	for i := range matrix {
		matrix[i] = make([]int, machines)
	}
	return matrix
}

func copyLinesOfFloatArray(arr [][]int, startRow, numLines int) [][]int {
	// Calculate the end row based on the number of lines to copy
	endRow := startRow + numLines

	// Check if endRow exceeds the array bounds
	if endRow > len(arr) {
		endRow = len(arr)
	}

	// Calculate the number of columns
	cols := len(arr[0])

	// Create a new 2D slice to hold the copied lines
	copied := make([][]int, numLines)
	for i := range copied {
		copied[i] = make([]int, cols)
	}

	// Copy elements from original array to the new subarray
	for i := startRow; i < endRow; i++ {
		copy(copied[i-startRow], arr[i])
	}

	return copied
}

func fillProbMatrix(probMatrix [][]int, individuals [][]int, numJobs int) [][]int {
	// Iterate over each individual
	for _, individual := range individuals {
		// Iterate over each job assignment in the individual
		for j := 0; j < len(individual); j++ {
			probMatrix[j][individual[j]]++
		}
	}
	// fmt.Print(probMatrix)
	// reader := bufio.NewReader(os.Stdin)
	// reader.ReadString('\n')
	return probMatrix
}

func selectWithChoice(weights []int) int {
	sum := 0
	for _, weight := range weights {
		sum += weight
	}

	random := rand.Intn(sum)
	cumulative := 0
	for i, weight := range weights {
		cumulative += weight
		if random < cumulative {
			return i
		}
	}
	return len(weights) - 1
}

func createNewIndividual(probMatrix [][]int, ET [][]float64) []int {
	newIndividual := make([]int, 0)

	for _, row := range probMatrix {
		choice := selectWithChoice(row)
		newIndividual = append(newIndividual, choice)
	}

	return newIndividual
}

func createNewPop(probMatrix [][]int, qtde int, ET [][]float64) [][]int {
	population := make([][]int, qtde)

	for i := 0; i < qtde; i++ {
		population[i] = createNewIndividual(probMatrix, ET)
	}

	return population
}

func sumTwoDArrays(arr1, arr2 [][]int) [][]int {
	// Check if arrays have the same dimensions
	if len(arr1) != len(arr2) || len(arr1[0]) != len(arr2[0]) {
		panic("Arrays must have the same dimensions to be summed.")
	}

	// Initialize result array with the same dimensions as arr1
	result := make([][]int, len(arr1))
	for i := range result {
		result[i] = make([]int, len(arr1[i]))
	}

	// Perform element-wise addition
	for i := 0; i < len(arr1); i++ {
		for j := 0; j < len(arr1[i]); j++ {
			result[i][j] = arr1[i][j] + arr2[i][j]
		}
	}

	return result
}

func appendTwoDArrays(arr1, arr2 [][]int) [][]int {
	// Check if arrays have the same number of columns
	if len(arr1[0]) != len(arr2[0]) {
		panic("Arrays must have the same number of columns to be appended.")
	}

	// Append each row of arr2 to arr1
	for _, row := range arr2 {
		arr1 = append(arr1, row)
	}

	return arr1
}

func mutate(individuals [][]int, mutProb int) [][]int {
	for _, i := range individuals {
		p := rand.Intn(100) + 1 // Generate random number from 1 to 100
		if p <= mutProb {
			pos := rand.Intn(len(i))          // Random position 1
			pos2 := rand.Intn(len(i))         // Random position 2
			i[pos], i[pos2] = i[pos2], i[pos] // Swap elements at pos and pos2
		}
	}

	return individuals
}

func getLastElement3D(arr3D [][][]int) int {
	last2D := arr3D[len(arr3D)-1]          // Get the last 2D array
	lastRow := last2D[len(last2D)-1]       // Get the last row in the last 2D array
	lastElement := lastRow[len(lastRow)-1] // Get the last element in the last row

	return lastElement
}

func saveToCSV(filename string, dict map[string]int, result float64) {
	// Open or create the CSV file
	file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the dictionary to the CSV file
	for key, value := range dict {
		record := []string{key, strconv.Itoa(value)}
		if err := writer.Write(record); err != nil {
			fmt.Printf("Failed to write record to file: %s", err)
		}
	}

	// Write the float result to the CSV file
	resultRecord := []string{"Result", fmt.Sprintf("%f", result)}
	if err := writer.Write(resultRecord); err != nil {
		fmt.Printf("Failed to write result to file: %s", err)
	}

	fmt.Printf("Data saved to CSV file successfully")
}

func main() {
	var creation_values = map[string][]int{
		"jobs":     {512},
		"machines": {16},
		"numInd":   {100, 200, 500},
		"numGen":   {100, 200, 500},
		"toMatrix": {10, 20, 30, 40, 50},
		"elitism":  {10, 20, 30, 40, 50},
		"mutate":   {10, 20, 30},
	}

	var pops [][][]int

	var keys []string
	for key := range creation_values {
		keys = append(keys, key)
	}

	best_makespan := 10000000000000.1

	var allCombinations []map[string]int
	generateCombinations(keys, creation_values, 0, make(map[string]int), &allCombinations)

	for i, combination := range allCombinations {
		fmt.Printf("Combination %d: %v\n", i+1, combination)
		var et_array = get_ET_as_array("512x16/u_c_hihi.0")
		var ET = generate_et_and_ct(et_array, combination["jobs"], combination["machines"])
		var CT [][]float64

		copy(CT, ET)

		var pop = get_n_random_elements(combination["jobs"], 0, combination["machines"]-1, combination["numInd"])
		pop = orderPop(ET, pop)
		pops = append(pops, pop)
		start := time.Now()

		for num_gen := 0; num_gen < combination["numGen"]; num_gen++ {
			pop = pops[len(pops)-1]
			pb := createProbMatrix(combination["jobs"], combination["machines"])

			to_matrix := copyLinesOfFloatArray(pop, 0, (combination["toMatrix"]*combination["numInd"])/100)

			pb = fillProbMatrix(pb, to_matrix, 0)
			pop := copyLinesOfFloatArray(pop, 0, combination["elitism"])

			new_pop := createNewPop(pb, combination["numInd"]-len(pop), ET)

			pop = appendTwoDArrays(pop, new_pop)

			pop = mutate(pop, combination["mutate"])
			pop = orderPop(ET, pop)
			pop = copyLinesOfFloatArray(pop, 0, combination["numInd"])
			// fmt.Println(getFitness(ET, pop[0]))

			// reader := bufio.NewReader(os.Stdin)
			// reader.ReadString('\n')

			pops = append(pops, pop)
		}

		pop = pops[len(pops)-1]

		fmt.Printf("%.6f ", getFitness(ET, pop[0]))
		elapsed := time.Since(start)
		fmt.Printf("Execution time: %.2f seconds\n", elapsed.Seconds())
		fmt.Printf("\n")

		saveToCSV("pureEDA.csv", combination, getFitness(ET, pop[0]))
	}

	fmt.Printf("%.6f ", best_makespan)

}

/*
// Print all combinations
	for i, combination := range allCombinations {
		fmt.Printf("Combination %d: %v\n", i+1, combination)
		fmt.Print(combination["numInd"])
	}
*/
