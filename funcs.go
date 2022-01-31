package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/lxn/walk"
	//. "github.com/lxn/walk/declarative"
)

// https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.7z

func createFfmpegFolder() {
	cmd := exec.Command("cmd", "/C", "mkdir", "ffmpeg")
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func isFfmpegInstalled() bool {
	args := []string{
		//"/C",
		//"WHERE.exe",
		//"/q",
		"ffmpeg.exe",
		//"2>&1",
	} //https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument
	cmd := exec.Command("where", args...)

	var serr bytes.Buffer
	cmd.Stderr = &serr
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	/*
		fmt.Println(cmd)
		//fmt.Println(out)
		//fmt.Println(out.String())
		fmt.Println("OUTPUCT:: ", cmd.Stdout)
		fmt.Println("OUTPUCT:: ", out.String())
		fmt.Println("ERRORR:", cmd.Stderr)
		fmt.Println("ERRORR:", serr.String())
		//fmt.Printf("%s", cmd)
		//fmt.Println(serr.String())
		//fmt.Println("CMD.OUT: ", string(out))

		fmt.Println("OUTPUCT:: ", reflect.TypeOf(cmd.Stdout))
		fmt.Println("OUTPUCT:: ", reflect.TypeOf(out.String()))
		fmt.Println("ERRORR:", reflect.TypeOf(cmd.Stderr))
		fmt.Println("ERRORR:", reflect.TypeOf(serr.String()))

		fmt.Println("LEN", len(out.String()))
		fmt.Println("LEN", len(serr.String()))
	*/
	if len(out.String()) == 0 {
		return false
	}

	return true
}

func openFileLocation(exPath, vidName string) {
	cmd := exec.Command(`explorer`, `/select,`+exPath+`\`+vidName+`.mp4`)
	fmt.Println(cmd)
	if err := cmd.Run(); err != nil {
		log.Println(err)
		fmt.Println("Ignore the error...")
	}
	// WHY THE ERROR!?!??!?!? It works
}

func removeThumbnail() {
	acmd := exec.Command("cmd", "/C", "rm", "./thumb.jpg")
	err := acmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func timeFormat(h, m, s string) string {
	if len(h) < 2 {
		h = addZeroes(h)
	}
	if len(m) < 2 {
		m = addZeroes(m)
	}
	if len(s) < 2 {
		s = addZeroes(s)
	}
	tt := h + m + s
	return tt
}

func addZeroes(t string) string {
	if len(t) == 0 {
		t = "00"
	} else if len(t) == 1 {
		t = "0" + t
	}
	return t
}

func formatStartTime(t string) string {
	th := t[0:2]
	tm := t[2:4]
	ts := t[4:6]
	tt := th + ":" + tm + ":" + ts
	return tt
}

func formatEndTime(s, e string) string {
	shh, err := strconv.Atoi(s[0:2])
	if err != nil {
		log.Fatal(err)
	}
	smm, err := strconv.Atoi(s[2:4])
	if err != nil {
		log.Fatal(err)
	}
	sss, err := strconv.Atoi(s[4:6])
	if err != nil {
		log.Fatal(err)
	}
	ehh, err := strconv.Atoi(e[0:2])
	if err != nil {
		log.Fatal(err)
	}
	emm, err := strconv.Atoi(e[2:4])
	if err != nil {
		log.Fatal(err)
	}
	ess, err := strconv.Atoi(e[4:6])
	if err != nil {
		log.Fatal(err)
	}

	tth, ttm, tts := convertFormat((ehh*3600 + emm*60 + ess) - (shh*3600 + smm*60 + sss))
	th := strconv.Itoa(tth)
	tm := strconv.Itoa(ttm)
	ts := strconv.Itoa(tts)
	tt := th + ":" + tm + ":" + ts
	return tt
}

func convertFormat(t int) (h, m, s int) {
	hh := t / 3600
	t = t - hh*3600
	mm := t / 60
	t = t - mm*60
	ss := t
	return hh, mm, ss
}

func fileExists(fp string) bool {
	info, err := os.Stat(fp)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getVideoDuration(path string) string {
	args := []string{
		"/C",
		`ffprobe.exe`,
		"-i",
		path,
		"-show_entries",
		"format=duration",
		"-v",
		"quiet",
		"-of",
		`csv=p=0`,
		"-sexagesimal",
	} //https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument
	// Don't use " " in csv="p=0"

	cmd := exec.Command(`cmd`, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return strings.Split(string(output), `.`)[0] //string(output)
}

func playVideo(item string) {
	args := []string{
		"/C",
		`ffplay.exe`,
		item,
		"-volume",
		"20",
		"-autoexit"} //https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument

	cmd := exec.Command("cmd", args...)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", cmd)
}

func getThumbnail(fp string) (walk.Image, error) {
	// ffmpeg -ss 00:00:01.00 -i input.mp4 -vf 'scale=320:320:force_original_aspect_ratio=decrease' -vframes 1 output.jpg
	// ffmpeg -i input.mp4 -vf  "thumbnail,scale=640:360" -frames:v 1 thumb.png
	args := []string{
		"/C",
		"rm",
		"thumb.jpg",
		"|",
		`ffmpeg.exe`,
		"-ss",
		"00:00:01.00",
		"-i",
		fp,
		"-vf",
		`scale=320:320:force_original_aspect_ratio=decrease`,
		"-vframes",
		"1",
		"thumb.jpg",
	} //https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument
	cmd := exec.Command("cmd", args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Printf("%s", cmd)
	defer removeThumbnail()
	return walk.NewImageFromFile("./thumb.jpg")
}

func removeIntermediates() {
	cmd := exec.Command(`cmd`)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.CmdLine = `/C del *.ts`
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
	}
}

/*

func (mw *MyMainWindow) playVideo(item string) {
	args := []string{
		"/C",
		mw.exPath + `\ffmpeg\ffplay.exe`,
		item,
		"-volume",
		"20",
		"-autoexit",
	} //https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument
	fmt.Println(mw.exPath)
	fmt.Println(item)

	cmd := exec.Command("cmd", args...)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", cmd)
}
*/
