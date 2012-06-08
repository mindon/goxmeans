/*
 Package xmeans implements a simple library for the xmeans algorithm.

 See Dan Pelleg and Andrew Moore: X-means: Extending K-means with Efficient Estimation of the Number of Clusters. 
*/
package xmeans

import (
	"fmt"
	"os"
	"bufio"
	"errors"
	"strconv"
	"strings"
	"io"
	"math"
	"math/rand"
	"code.google.com/p/gomatrix/matrix"
	"matutil"
)

// Atof64 is shorthand for ParseFloat(s, 64)
func Atof64(s string) (f float64, err error) {
	f64, err := strconv.ParseFloat(s, 64)
	return float64(f64), err
}

	
// Load loads a tab delimited text file of floats into a slice.
// Assume last column is the target.
// For now, we limit ourselves to two columns
func Load(fname string) (*matrix.DenseMatrix, error)  {
	datamatrix := matrix.Zeros(1, 1);
	data := make([]float64, 2048) 

	fp, err := os.Open(fname)
	if err != nil {
		return datamatrix, err
	}
	defer fp.Close()

	r := bufio.NewReader(fp)
	linenum := 1
	eof := false
	for !eof {
		var line string
		line, err := r.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
			break
		} else 	if err != nil {
			return datamatrix, errors.New(fmt.Sprintf("means: reading linenum %d: %v", linenum, err))
		} 
//		fmt.Printf("debug: linenum=%d line=%s\n", linenum, line)

		linenum++
		l1 := strings.TrimRight(line, "\n")
		l := strings.Split(l1, "\t")
		if len(l) < 2 {
			return datamatrix, errors.New(fmt.Sprintf("means: linenum %d has only %d elements", linenum, len(line)))
		}

		// for now assume 2 dimensions only
		f0, err := Atof64(string(l[0]))
		if err != nil {
			return datamatrix, errors.New(fmt.Sprintf("means: cannot convert %s to float64.", l[0]))
		}
		f1, err := Atof64(string(l[1]))
		if err != nil {
			return datamatrix, errors.New(fmt.Sprintf("means: cannot convert %s to float64.", l[linenum][1]))
		}
		data = append(data, f0, f1)
	}
	numcols := 2
	datamatrix = matrix.MakeDenseMatrix(data, len(data)/numcols, numcols)
	return datamatrix, nil
}

// randcent picks random centroids 
func RandCentroids(mat *matrix.DenseMatrix, k int) (matrix.DenseMatrix) {
	rows,cols := mat.GetSize()
	minj := 0
	for j := 0; j <  cols; j++ {
		r := matutil.ColSlice(mat, j)
		// min value from each column
		for _, val := range r {
			minj := math.Min(minj, val)
		}

		// max value from each column
		maxj := 0
		for _,val = range r {
			maxj := math.Max(maxj, val)
		}

		// create a slice of random centroids 
		// base on minj + rangeJ * random num based on k
		rands := make([]float64, rows)
		for i := 0; i < rows; i++ {
			rands = appeand(rands,  maxj - minj * rand.Float64())
		}

		centroids, err := AppendCol(mat, rands)
		if err != nil {
			return centroids, errors.New(fmt.Sprintf("means: RandCentroids could not append column.  err=%v", err))
		}
	}
	return centroids
}
