package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	stdHor = 1920
	stdVer = 1080
)

type image struct {
	fname       string
	format      string
	orientation string
	hSize       int
	vSize       int
}

func main() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	info := []string{}
	for _, f := range files {
		cmd := exec.Command("file", f.Name())
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Fatal("error runing file command ", err)
		}

		info = append(info, out.String())
	}

	// tyssul-patel-o-zOatT4kQw-unsplash.jpg: JPEG image data, JFIF standard 1.01, resolution (DPI), density 72x72, segment length 16, baseline, precision 8, 4896x3264, components 3
	for _, f := range info {
		// match 5323x8293, 934x482, and others
		re := regexp.MustCompile("\\d{3,5}x\\d{3,5}")
		size := re.FindString(f)

		a := getCharacteristics(f)
		fi := isImage(a)
		img := image{}
		if fi {
			img = getImageDetails(a, size)
			convertImage(img)
		}
		// parenthesis because of some parsing ambiguity
		if img != (image{}) {
			convertImage(img)
		}
	}
}

// getCharacteristics builds slice with the fields
// corresponding to the `file` command output
func getCharacteristics(s string) []string {
	fields := strings.Split(s, ",")
	els := []string{}
	for i, v := range fields {
		if i == 0 {
			nametype := strings.Split(v, ": ")
			els = append(els, nametype...)
		} else {
			els = append(els, strings.Trim(v, " \n"))
		}
	}
	return els
}

func isImage(f []string) bool {
	img := false
	if f[1] == "JPEG image data" {
		img = true
	}
	return img
}

func getImageDetails(a []string, sz string) image {
	img := image{}
	img.fname = a[0]
	img.format = a[1]
	d := getSize(sz)

	img.hSize = d[0]

	img.vSize = d[1]
	if img.hSize >= img.vSize {
		img.orientation = "landscape"
	} else {
		img.orientation = "portrait"
	}
	return img
}

func getSize(s string) []int {
	sizes := []int{}
	ssplit := strings.Split(s, "x")

	sizeh, err := strconv.Atoi(ssplit[0])
	if err != nil {
		log.Fatal("not possible to convert to integer", ssplit[0])
	}
	sizes = append(sizes, sizeh)

	sizev := 0
	sizev, err = strconv.Atoi(ssplit[1])
	if err != nil {
		log.Fatal("not possible to convert to integer", ssplit[1])
	}
	sizes = append(sizes, sizev)

	return sizes
}

func convertImage(img image) {
	hRot := 0
	vRot := 0
	if (img.hSize >= img.vSize && stdHor >= stdVer) || (img.hSize < img.vSize && stdHor < stdVer) {
		// image it's in landscape mode as the model wanted
		hRot = img.hSize
		vRot = img.vSize
	} else {
		hRot = img.vSize
		vRot = img.hSize
	}

	hRatio := float32(stdHor) / float32(hRot)
	vRatio := float32(stdVer) / float32(vRot)
	mRatio := (hRatio + vRatio) / 2
	mRatioStr := strconv.Itoa(int(mRatio*100)) + "%"

	convertName := newName(img)

	fmt.Println(convertName, mRatioStr)
	cmd := exec.Command("convert", img.fname, "-resize", mRatioStr, convertName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal("error converting image ", err, " > ", stderr.String())
	}
	fmt.Println(out.String())
}

func newName(img image) string {
	i := strings.LastIndex(img.fname, ".")
	n := img.fname[:i]
	n = n + "_0"
	if img.format == "JPEG image data" {
		n = n + ".jpg"
	}
	return n
}
