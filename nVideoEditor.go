package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyMainWindow struct {
	*walk.MainWindow
	exPath       string
	textEdit     *walk.TextEdit
	nameEdit     *walk.TextEdit
	prevFilePath string
	vidName      string
	vidLength    *walk.Label
	imgName      string
	ni           *walk.NotifyIcon
	thumbIV      *walk.ImageView
}

func assertErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	// Create a new MainWindow
	mw := new(MyMainWindow)

	ex, err := os.Executable()
	assertErr(err)
	mw.exPath = filepath.Dir(ex)

	var startEdith, startEditm, startEdits, endEdith, endEditm, endEdits *walk.TextEdit
	times := struct{ Sh, Sm, Ss, Eh, Em, Es string }{"", "", "", "", "", ""}
	mw.vidName = "cut Video Name"
	//	timg, _ := walk.Resources.Image("./ico256.ico")

	if err := (MainWindow{
		Size:     Size{400, 500},
		AssignTo: &mw.MainWindow,
		Title:    "Nut Video Editor",
		OnDropFiles: func(files []string) {
			//mw.imageView.Image().Dispose()
			mw.textEdit.SetText(strings.Join(files, "\r\n"))
			filepa := files[0]
			fs := strings.SplitAfter(filepa, `\`)
			fn := fs[len(fs)-1]
			filename := string(fn)[:len(string(fn))-4]
			mw.nameEdit.SetText(filename)
			mw.prevFilePath = string(files[0])
			mw.vidLength.SetText(getVideoDuration(string(files[0]), mw.exPath))
			mw.getThumbnail(filepa)
		},
		Layout: VBox{Margins: Margins{10, 10, 10, 10}, Spacing: 10},
		Children: []Widget{
			TextEdit{
				CompactHeight: true,
				MaxSize:       Size{100, 100},
				AssignTo:      &mw.textEdit,
				ReadOnly:      true,
				Text:          "Drop a video here, from windows explorer...",
			},
			PushButton{
				Text:      "Select Video",
				OnClicked: mw.openAction_Triggered,
			},
			Label{Text: "Video Length"},
			Label{
				AssignTo: &mw.vidLength,
				Text:     "",
			},
			ImageView{
				AssignTo: &mw.thumbIV,
				//Image:    timg,
				Margin: 10,
				Mode:   ImageViewModeZoom,
			},
			Composite{
				Layout: Grid{
					Columns: 5,
				},
				StretchFactor: 4,
				Children: []Widget{
					Label{
						Text:       "",
						ColumnSpan: 1,
					},
					Label{
						Text:       "hh",
						ColumnSpan: 1,
					},
					Label{
						Text:       "mm",
						ColumnSpan: 1,
					},
					Label{
						Text:       "ss",
						ColumnSpan: 1,
					},
					Label{
						Text:       "",
						ColumnSpan: 1,
					},
					Label{
						Text:       "Start Time",
						ColumnSpan: 1,
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &startEdith,
						Text:          times.Sh,
						OnTextChanged: func() {
							times.Sh = startEdith.Text()
						},
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &startEditm,
						Text:          times.Sm,
						OnTextChanged: func() {
							times.Sm = startEditm.Text()
						},
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &startEdits,
						Text:          times.Ss,
						OnTextChanged: func() {
							times.Ss = startEdits.Text()
						},
					},
					Label{
						Text:       "",
						ColumnSpan: 1,
					},
					Label{
						Text:       "End Time",
						ColumnSpan: 1,
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &endEdith,
						Text:          times.Eh,
						OnTextChanged: func() {
							times.Eh = endEdith.Text()
						},
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &endEditm,
						Text:          times.Em,
						OnTextChanged: func() {
							times.Em = endEditm.Text()
						},
					},
					TextEdit{
						MaxLength:     2,
						CompactHeight: true,
						ColumnSpan:    1,
						MaxSize:       Size{30, 30},
						AssignTo:      &endEdits,
						Text:          times.Es,
						OnTextChanged: func() {
							times.Es = endEdits.Text()
						},
					},
					PushButton{
						Text:       "To end of of the video",
						ColumnSpan: 1,
						OnClicked: func() {
							if len(mw.prevFilePath) > 1 {
								time := getVideoDuration(mw.prevFilePath, mw.exPath)
								times.Eh = time[:1]
								times.Em = time[2:4]
								times.Es = time[5:7]
								endEdith.SetText(addZeroes(times.Eh))
								endEditm.SetText(times.Em)
								endEdits.SetText(times.Es)
							}
						},
					},
				},
			},

			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						ColumnSpan: 2,
						Text:       "Name of the cut video",
					},
					TextEdit{
						CompactHeight: true,
						AssignTo:      &mw.nameEdit,
						Text:          mw.vidName,
						OnTextChanged: func() {
							mw.vidName = mw.nameEdit.Text()
						},
					},
					TextLabel{Text: ".mp4"},
				},
			},
			PushButton{
				Text: "Cut",
				OnClicked: func() {
					if mw.prevFilePath == "" {
						walk.MsgBox(mw, "Error", "No video selected", walk.MsgBoxIconInformation)
					} else {
						mw.cutVideo(mw.exPath, mw.prevFilePath, mw.vidName, times.Sh, times.Sm, times.Ss, times.Eh, times.Em, times.Es)
					}
				},
			},
			PushButton{
				Text: "Open File Location",
				OnClicked: func() {
					cmd := exec.Command(`explorer`, `/select,`, mw.exPath+`\`+mw.vidName+`.mp4`)
					fmt.Println(cmd)
					if err := cmd.Run(); err != nil {
						log.Println(err)
					}
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	mw.initNotifyIcon()
	defer mw.ni.Dispose()

	mw.Run()
}

func (mw *MyMainWindow) initNotifyIcon() {
	// Create the notify icon and make sure we clean it up on exit.
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	// We load our icon from a file.
	if fileExists("./ico256.ico") {
		icon, err := walk.Resources.Icon("./ico256.ico")
		if err != nil {
			log.Fatal(err)
		}
		// Set the icon and a tool tip text.
		if err := mw.ni.SetIcon(icon); err != nil {
			log.Fatal(err)
		}
	}
	if err := mw.ni.SetToolTip("Click for info or use the context menu to exit."); err != nil {
		log.Fatal(err)
	}
	// When the left mouse button is pressed, bring up our balloon.
	mw.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.LeftButton {
			return
		} /*
			if err := mw.ni.ShowCustom(
				"Walk NotifyIcon Example",
				"There are multiple ShowX methods sporting different icons.",
				icon); err != nil {
				log.Fatal(err)
			} */
	})
	// We put an exit action into the context menu.
	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := mw.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}
	// The notify icon is hidden initially, so we have to make it visible.
	if err := mw.ni.SetVisible(true); err != nil {
		log.Fatal(err)
	}
}

func (mw *MyMainWindow) openAction_Triggered() {
	if err := mw.openFile(); err != nil {
		log.Print(err)
	}
}

func (mw *MyMainWindow) openFile() error {

	dlg := new(walk.FileDialog)
	dlg.FilePath = mw.prevFilePath
	//dlg.Filter = "Image Files (*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff)|*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff"
	dlg.Filter = "Image Files (*.mp4;*.mkv)|*.mp4;*.mkv"
	dlg.Title = "Select a Video"

	if ok, err := dlg.ShowOpen(mw); err != nil {
		return err
	} else if !ok {
		return nil
	}

	mw.prevFilePath = dlg.FilePath
	mw.textEdit.SetText(dlg.FilePath)

	openedFile, err := os.Stat(dlg.FilePath)
	if err != nil {
		log.Println(err)
	}

	// set filename without extension
	filename := openedFile.Name()[:len(openedFile.Name())-4]
	mw.nameEdit.SetText(filename)

	// show duration of selected video
	mw.vidLength.SetText(getVideoDuration(dlg.FilePath, mw.exPath))

	mw.getThumbnail(mw.prevFilePath) //____________________________________________________________________

	/* ImageView{
		Image:  timg,
		Margin: 10,
		Mode:   ImageViewModeZoom,
	},
	*/
	return nil
}

func getVideoDuration(path, exPath string) string {
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
	}
	// Don't use " " in csv="p=0"

	cmd := exec.Command(`cmd`, args...)
	cmd.Dir = exPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return strings.Split(string(output), `.`)[0] //string(output)
}

// FHHHUuuuhHHUH arguments are formatted automatically probably. Took me a while to understand that
func (mw *MyMainWindow) cutVideo(exPath, item, name, sh, sm, ss, eh, em, es string) {
	if fileExists(exPath + `\` + name + ".mp4") {
		timenow := time.Now()
		name = name + "-" + timeFormat(strconv.Itoa(timenow.Hour()), strconv.Itoa(timenow.Minute()), strconv.Itoa(timenow.Second()))
	}
	start := timeFormat(sh, sm, ss)
	end := timeFormat(eh, em, es)
	args := []string{
		"/C",
		`ffmpeg.exe`,
		"-ss",
		formatStartTime(start),
		"-i",
		item,
		"-to",
		formatEndTime(start, end),
		"-c",
		"copy",
		"-f",
		"mp4",
		name + ".mp4",
	}

	cmd := exec.Command(`cmd`, args...)
	cmd.Dir = exPath
	//var stderr bytes.Buffer
	//cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cmd.Stdout)
	// Now that the icon is visible, we can bring up an info balloon.
	if err := mw.ni.ShowInfo("Video cutting done!", "Thanks for using Nve!"); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(stderr)
	// https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
	fmt.Printf("%s", cmd)
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

func (mw *MyMainWindow) getThumbnail(fp string) {
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
	}
	cmd := exec.Command("cmd", args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Printf("%s", cmd)

	ivm, _ := walk.NewImageFromFile("./thumb.jpg")
	mw.thumbIV.SetImage(ivm)

	removeThumbnail()
}

func removeThumbnail() {
	acmd := exec.Command("cmd", "/C", "rm", "./thumb.jpg")
	err := acmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func (mw *MyMainWindow) playVideo(item string) {
	args := []string{
		"/C",
		mw.exPath + `\ffmpeg\ffplay.exe`,
		item,
		"-volume",
		"20",
		"-autoexit"}
	fmt.Println(mw.exPath)
	fmt.Println(item)

	cmd := exec.Command("cmd", args...)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", cmd)
}
