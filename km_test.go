package goxmeans

import (
	"bufio"
//	"code.google.com/p/gomatrix/matrix"
	"fmt"
	"os"
//	"math"
	"testing"
	"github.com/bobhancock/gomatrix/matrix"
	"goxmeans/matutil"
)

var DATAPOINTS = matrix.MakeDenseMatrix([]float64{3.275154,2.957587,
	-3.344465,2.603513,
	0.355083,-3.376585,
	1.852435,3.547351,
	-2.078973,2.552013,
	-0.993756,-0.884433,
	2.682252,4.007573,
	-3.087776,2.878713,
	-1.565978,-1.256985,
	2.441611,0.444826,
	103.29,209.6594,
	125.93,230.3988}, 12, 2)

var CENTROIDS = matrix.MakeDenseMatrix([]float64{ 46.57890839,   11.95938243,
    67.54513486,  134.19858589,
    81.16283573,   95.83181046}, 3, 2)

func TestAtof64Invalid(t *testing.T) {
	s := "xyz"
	if _, err := Atof64(s); err == nil {
		t.Errorf("err == nil with invalid input %s.", s)
	}
}

func TestAtof64Valid(t *testing.T) {
	s := "1234.5678"
	if f64, err := Atof64(s); err != nil {
		t.Errorf("err != nil with input %s. Returned f64=%f,err= %v.", s, f64, err)
	}
}

func TestFileNotExistsLoad(t *testing.T) {
	f := "filedoesnotexist"
	if _, err := Load(f); err == nil {
		t.Errorf("err == nil with file that does not exist.  err=%v.", err)
	}
}

func createtestfile(fname, record string) (int, error) {
	fp, err := os.Create(fname)
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	w := bufio.NewWriter(fp)
	i, err := w.WriteString(record)
	if err != nil {
		return i, err
	}
	w.Flush()

	return i, err
}

// Does the input line contain < 2 elements
func TestInputInvalid(t *testing.T) {
	fname := "inputinvalid"
	_, err := createtestfile(fname, "123\n")
	if err != nil {
		t.Errorf("Could not create test file. err=%v", err)
	}
	defer os.Remove(fname)

	if _, err := Load(fname); err == nil {
		t.Errorf("err: %v", err)
	}
}

func TestValidReturnLoad(t *testing.T) {
	fname := "inputvalid"
	record := fmt.Sprintf("123\t456\n789\t101")
	_, err := createtestfile(fname, record)
	if err != nil {
		t.Errorf("Could not create test file %s err=%v", fname, err)
	}
	defer os.Remove(fname)

	if _, err := Load(fname); err != nil {
		t.Errorf("Load(%s) err=%v", fname, err)
	}
}

/* Test fails
func TestRandCentroids(t *testing.T) {
	rows := 3
	cols := 3
	k := 2
	data := []float64{1, 2.0, 3, -4.945, 5, -6.1, 7, 8, 9}
	mat := matrix.MakeDenseMatrix(data, rows, cols)
	choosers := []CentroidChooser{RandCentroids{}, DataCentroids{}, EllipseCentroids{0.5}}
	for _, cc := range choosers{
		centroids := cc.ChooseCentroids(mat, k)

		r, c := centroids.GetSize()
		if r != k || c != cols {
			t.Errorf("Returned centroid was %dx%d instead of %dx%d", r, c, rows, cols)
		}
	}
}
*/


func TestComputeCentroid(t *testing.T) {
	empty := matrix.Zeros(0, 0)
	_, err := ComputeCentroid(empty)
	if err == nil {
		t.Errorf("Did not raise error on empty matrix")
	}
	twoByTwo := matrix.Ones(2, 2)
	centr, err := ComputeCentroid(twoByTwo)
	if err != nil {
		t.Errorf("Could not compute centroid, err=%v", err)
	}
	expected := matrix.MakeDenseMatrix([]float64{1.0, 1.0}, 1, 2)
	if !matrix.Equals(centr, expected) {
		t.Errorf("Incorrect centroid: was %v, should have been %v", expected, centr)
	}
	twoByTwo.Set(0, 0, 3.0)
	expected.Set(0, 0, 2.0)
	centr, err = ComputeCentroid(twoByTwo)
	if err != nil {
		t.Errorf("Could not compute centroid, err=%v", err)
	}
	if !matrix.Equals(centr, expected) {
		t.Errorf("Incorrect centroid: was %v, should have been %v", expected, centr)
	}
}


func TestKmeansp(t *testing.T) {
	dataPoints, err := Load("./testSetSmall.txt")
	if err != nil {
		t.Errorf("Load returned: %v", err)
		return
	}
	
	var ed matutil.EuclidDist
	var cc RandCentroids
	//centroidsdata := []float64{1.5,1.5,2,2,3,3,0.9,0,9}
	//centroids := matrix.MakeDenseMatrix(centroidsdata, 4,2)

	centroidMeans, centroidSqDist, err := Kmeansp(dataPoints, 4, cc, ed)
	if err != nil {
		t.Errorf("Kmeans returned: %v", err)
		return
	}

	if 	a, b := centroidMeans.GetSize(); a == 0 || b == 0 {
		t.Errorf("Kmeans centroidMeans is of size %d, %d.", a,b)
	}

	if c, d := centroidSqDist.GetSize(); c == 0 || d == 0 {
		t.Errorf("Kmeans centroidSqDist is of size %d, %d.", c,d)
	}
}
   
func TestAddPairPointToCentroidJob(t *testing.T) {
	r := 4
	c := 2
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))
	dataPoints := matrix.Zeros(r, c)
	centroidSqDist := matrix.Zeros(r, c)
	centroids := matrix.Zeros(r, c)

	var ed matutil.EuclidDist
	
	go addPairPointCentroidJobs(jobs, dataPoints, centroids, centroidSqDist,ed ,results)
	i := 0
	for ; i < r; i++ {
        <-jobs 
		//fmt.Printf("Drained %d\n", i)
    }

	if i  != r {
		t.Errorf("addPairPointToCentroidJobs number of jobs=%d.  Should be %d", i, r)
	}
}
	
func TestDoPairPointCentroidJobs(t *testing.T) {
	r := 4
	c := 2
	dataPoints := matrix.Zeros(r, c)
	centroidSqDist := matrix.Zeros(r, c)
	centroids := matrix.Zeros(r, c)

	done := make(chan int)
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))

	var md matutil.ManhattanDist

	go addPairPointCentroidJobs(jobs, dataPoints, centroids, centroidSqDist, md, results)

	for i := 0; i < r; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}

	j := 0
	for ; j < r; j++ {
        <- done
    }

	if j  != r {
		t.Errorf("doPairPointToCentroidJobs jobs processed=%d.  Should be %d", j, r)
	}
}

func TestAssessClusters(t *testing.T) {
	r, c := DATAPOINTS.GetSize()
	clusterAssessment := matrix.Zeros(r, c)

	done := make(chan int)
	jobs := make(chan PairPointCentroidJob, r)
	results := make(chan PairPointCentroidResult, minimum(1024, r))

	var md matutil.ManhattanDist
	go addPairPointCentroidJobs(jobs, DATAPOINTS, CENTROIDS, clusterAssessment, md, results)

	for i := 0; i < r; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}
	go awaitPairPointCentroidCompletion(done, results)

    //TODO check deterministic results of clusterAssessment
    clusterChanged := assessClusters(clusterAssessment, results)
	if clusterChanged != true {
		t.Errorf("TestAssessClusters: clusterChanged=%b and should be true.", clusterChanged)
	}


// clusterAssessment should be:
//           {0,  2735.870542,
//            0,  3514.028629,
//           0,  3789.608092,
//            0,  2823.700695,
//            0,  3371.573353,
//            0,  3650.151034,
//            0,  2688.263408,
//            0,  3451.251581,
//           0,   3765.20347,
//            0,  3097.128834,
//            1, 12366.703097,
//            1, 23896.546727}


	if clusterAssessment.Get(9, 0) != 0 || clusterAssessment.Get(10, 0) != 1 {
		t.Errorf("TestAssessClusters: rows 9 and 10 should have 0 and 1 in column 0, but received %v", clusterAssessment)
	}
}

/* TODO rewrite for new version
func TestKmeansbi(t *testing.T) {
	var ed matutil.EuclidDist
	var cc RandCentroids

	matCentroidlist, clusterAssignment, err := Kmeansp(DATAPOINTS, 4, cc, ed)
	if err != nil {
		t.Errorf("Kmeans returned: %v", err)
		return
	}

	if 	a, b := matCentroidlist.GetSize(); a == 0 || b == 0 {
		t.Errorf("Kmeans centroidMeans is of size %d, %d.", a,b)
	}

	if c, d := clusterAssignment.GetSize(); c == 0 || d == 0 {
		t.Errorf("Kmeans clusterAssessment is of size %d, %d.", c,d)
	}
	// TODO deterministic test
}
*/
/*  
func TestVariance(t *testing.T) {
	numRows, numCols := DATAPOINTS.GetSize()
	clusterAssessment := matrix.Zeros(numRows, numCols)

	jobs := make(chan PairPointCentroidJob, numworkers)
	results := make(chan PairPointCentroidResult, minimum(1024, numRows))
	done := make(chan int, numworkers)
	var measurer matutil.EuclidDist 

	go addPairPointCentroidJobs(jobs, DATAPOINTS, clusterAssessment, CENTROIDS, measurer, results)
	for i := 0; i < numworkers; i++ {
		go doPairPointCentroidJobs(done, jobs)
	}
	go awaitPairPointCentroidCompletion(done, results)
	b := assessClusters(clusterAssessment, results) // This blocks so that all the results can be processed
	fmt.Println(b, clusterAssessment)

	v, err := variance(DATAPOINTS, CENTROIDS, clusterAssessment, 1, measurer)
	if err != nil {
		t.Errorf("TestVariance: err = %v", err)
	}
	
	E := 4.000000
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(v - na) 

	if diff > epsilon {
		t.Errorf("TestVariance: excpected %f but received %f.  The difference %f exceeds epsilon %f.", E, v, diff, epsilon)
	}
}
*/
/*
func TestPointProb(t *testing.T) {
	R := 10010.0
	Ri := 100.0
	M := 2.0
	V := 20.000000

	point := matrix.MakeDenseMatrix([]float64{5, 7},
		1,2)

	mean := matrix.MakeDenseMatrix([]float64{6, 8},
		1,2)

	var ed matutil.EuclidDist 

	//	pointProb(R, Ri, M int, V float64, point, mean *matrix.DenseMatrix, measurer matutil.VectorMeasurer) (float64, error) 
	pp := pointProb(R, Ri, M, V, point, mean, ed)

	E :=  0.011473
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(pp - na) 

	if diff > epsilon {
		t.Errorf("TestPointProb: expected %f but received %f.  The difference %f exceeds epsilon %f", E, pp, diff, epsilon)
	}
}

/*func TestLogLikeli(t *testing.T) {
	// TODO In Progress
	K := 5.0  // 5 clusters
	M := 2.0 // Dimensions

	D, err := Load("./testSet.txt")
	if err != nil {
		t.Errorf("TestLogLikeli: Load returned err=%s.", err)
	}

	var ed matutil.EuclidDist
	r, _ := D.GetSize()
	R := float64(r)
	mean := D.MeanCols()
	V := variance(D, mean, K, ed)
	Rn := []float64{R}

	ll := loglikeli(R, M, V, K, Rn)

	E :=  422.331625
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(ll - na) 

	if diff > epsilon {
		t.Errorf("TestLoglikeli: Expected 422.331625 but received %f.  The difference %f exceeds epsilon %f", E, ll, diff, epsilon)
	}
}

func TestFreeparams(t *testing.T) {
	K := 6.0
	M := 3.0

	r := freeparams(K, M)
	if r != 24. {
		t.Errorf("TestFreeparams: Expected 24 but received %f.", r)
	}
}

func TestBIC(t *testing.T) {
	K := 6.0
	M := 3.0
	fp := freeparams(K, M)
	loglike := 3199.331
	R := 1000.0
	bic := BIC(loglike, fp, R)
	
	E :=  3180.423244721018
	epsilon := .000001
	na := math.Nextafter(E, E + 1) 
	diff := math.Abs(bic - na) 
	if diff > epsilon {
		t.Errorf("TestBIC: Expected %f but received %f.  The difference %f exceeds epsilon %f", E, bic, diff, epsilon)
	}
}
*/