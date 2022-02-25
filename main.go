package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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

	vidList     []string
	vidListComp *walk.Composite

	cutAttributes CutAttributes // Tästä vois jatkaa parantamalla tätä

	cutPrevFilePath string
	cutTextEdit     *walk.TextEdit
	cutNameEdit     *walk.TextEdit
	cutVidName      string
	cutThumbIV      *walk.ImageView
	cutTextLabel    *walk.TextLabel

	conPrevFilePath string
	conTextEdit     *walk.TextEdit
	conNameEdit     *walk.TextEdit
	conVidName      string
	conThumbIV      *walk.ImageView
	conWorking      bool

	progressComposite     *walk.Composite
	progressBar           *walk.CustomWidget
	progressBarLabel      *walk.TextLabel
	progressFullLength    string
	progressCurrentLength string
}

type CutAttributes struct {
	cutPrevFilePath string
	cutTextEdit     *walk.TextEdit
	cutNameEdit     *walk.TextEdit
	cutVidName      string
	cutThumbIV      *walk.ImageView
	cutTextLabel    *walk.TextLabel
}

// there must be a better way to do this :>
type CutComposite struct {
	*walk.Composite
}

type ConcatComposite struct {
	*walk.Composite
}

var ffmpegPath string
var ffprobePath string

func main() {

	// Create a new MainWindow
	mw := new(MyMainWindow)
	cut := new(CutComposite)
	con := new(ConcatComposite)

	ex, err := os.Executable()
	if err != nil {
		log.Panic("asdasd", err)
	}
	mw.exPath = filepath.Dir(ex)

	var startEdith, startEditm, startEdits, endEdith, endEditm, endEdits *walk.TextEdit
	times := struct{ Sh, Sm, Ss, Eh, Em, Es string }{"", "", "", "", "", ""}
	mw.vidName = "cut Video Name"
	mw.cutVidName = "cut"
	mw.conWorking = false
	//	timg, _ := walk.Resources.Image("./ico256.ico")

	if err := (MainWindow{
		Size:     Size{Width: 360, Height: 540},
		AssignTo: &mw.MainWindow,
		Title:    "Nut Video Editor",
		MenuItems: []MenuItem{
			Action{
				Text: "&Cut",
				OnTriggered: func() {
					cut.SetVisible(true)
					cut.SetEnabled(true)
					con.SetVisible(false)
					con.SetEnabled(false)
				},
				//Image: "../img/document-new.png",
			},
			Separator{},
			Action{
				Text: "Conc&at",
				OnTriggered: func() {
					cut.SetVisible(false)
					cut.SetEnabled(false)
					con.SetVisible(true)
					con.SetEnabled(true)
				},
				//Image: "../img/document-properties.png",
			},
			Separator{},
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{
						Text:        "&Help",
						OnTriggered: func() { mw.helpAction_Triggered() },
					},
					Action{
						Text:        "&About",
						OnTriggered: func() { mw.aboutAction_Triggered() },
					},
				},
			},
		},
		OnDropFiles: func(files []string) {
			//mw.openVideo(string(files[0])) // if edit or concat visible/enable
			fmt.Println(string(files[0])[len(string(files[0]))-4:])
			if string(files[0])[len(string(files[0]))-4:] == ".mp4" || string(files[0])[len(string(files[0]))-4:] == ".mkv" || string(files[0])[len(string(files[0]))-4:] == ".wmv" {
				if cut.Enabled() {
					mw.cutSetVideo(string(files[0]))
				}
				if con.Enabled() {
					mw.concatSetVideo(string(files[0]))
				}
			} else {
				walk.MsgBox(mw,
					"Wrong type of file",
					"Added file needs to be in - .mp4 | .mkv | .wmv - format for now",
					walk.MsgBoxOK|walk.MsgBoxIconInformation)
			}
		},
		Layout: VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}, Spacing: 10},
		Children: []Widget{
			Composite{
				AssignTo: &cut.Composite,
				Visible:  true,
				Enabled:  true,
				Layout:   VBox{MarginsZero: true, SpacingZero: true},
				Children: []Widget{
					TextLabel{
						//CompactHeight: true,
						MaxSize:  Size{Width: 100, Height: 100},
						AssignTo: &mw.cutTextLabel,
						//ReadOnly:      true,
						Text: "Drop a video here",
					},
					PushButton{
						Text: "Select Video",
						OnClicked: func() {
							mw.cutOpenFile()
						},
					},
					Composite{ // HSpacer{} !!__!!_!__!_!__!_!_!_!__!_!_!__!
						Layout: Grid{
							Columns:     3,
							MarginsZero: true,
							SpacingZero: true,
						},
						Children: []Widget{
							Label{
								Text:       "Video Length: ",
								MinSize:    Size{Width: 100, Height: 0},
								ColumnSpan: 1,
							},
							Label{
								AssignTo:   &mw.vidLength,
								Text:       " ",
								MinSize:    Size{Width: 100, Height: 0},
								ColumnSpan: 1,
							},
							HSpacer{
								//Text:       " ",
								MinSize:    Size{Width: 50, Height: 0},
								ColumnSpan: 1,
							},
						},
					},
					ImageView{
						AssignTo: &mw.cutThumbIV,
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
								MaxSize:       Size{Width: 30, Height: 30},
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
								MaxSize:       Size{Width: 30, Height: 30},
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
								MaxSize:       Size{Width: 30, Height: 30},
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
								MaxSize:       Size{Width: 30, Height: 30},
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
								MaxSize:       Size{Width: 30, Height: 30},
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
								MaxSize:       Size{Width: 30, Height: 30},
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
									if len(mw.cutPrevFilePath) > 1 {
										time := getVideoDuration(mw.cutPrevFilePath)
										getVideoDurationMM(mw.cutPrevFilePath)
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
								AssignTo:      &mw.cutNameEdit,
								Text:          mw.cutVidName,
								OnTextChanged: func() {
									mw.cutVidName = mw.cutNameEdit.Text()
								},
							},
							TextLabel{Text: ".mp4"},
						},
					},
					PushButton{
						Text: "Cut",
						OnClicked: func() {
							if mw.cutPrevFilePath == "" {
								walk.MsgBox(mw, "Error", "No video selected", walk.MsgBoxIconInformation)
							} else {
								go mw.cutVideo(mw.exPath, mw.cutPrevFilePath, mw.cutVidName, times.Sh, times.Sm, times.Ss, times.Eh, times.Em, times.Es, *mw.ni)

								//mw.cutVideo(mw.exPath, mw.prevFilePath, mw.cutVidName, times.Sh, times.Sm, times.Ss, times.Eh, times.Em, times.Es)
							}
						},
					},
					PushButton{
						Text: "Open File Location",
						OnClicked: func() {
							fmt.Println("PUSHED OPEN LOACAION")
							openFileLocation(mw.exPath, mw.cutVidName)
							/* cmd := exec.Command(`explorer`, `/select,`, mw.exPath+`\`+mw.cutVidName+`.mp4`)
							fmt.Println(cmd)
							if err := cmd.Run(); err != nil {
								log.Println(err)
							} */
						},
					},
				},
			},
			Composite{
				AssignTo: &con.Composite,
				Layout:   VBox{Margins: Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}, Spacing: 5},
				Visible:  false,
				Enabled:  false,
				Children: []Widget{
					TextEdit{
						CompactHeight: true,
						//MaxSize:       Size{100, 100},
						AssignTo: &mw.conTextEdit,
						ReadOnly: true,
						Text:     "Drop a video here",
					},
					PushButton{
						Text: "Add Video",
						OnClicked: func() {
							mw.concatOpenFile()
						},
					},
					Composite{
						MinSize:            Size{Width: 300, Height: 300},
						AlwaysConsumeSpace: true,
						AssignTo:           &mw.vidListComp,
						Layout: HBox{
							Margins: Margins{Left: 10, Top: 0, Right: 10, Bottom: 0},
						},
						Children: []Widget{},
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
								AssignTo:      &mw.conNameEdit,
								Text:          mw.conVidName,
								OnTextChanged: func() {
									mw.conVidName = mw.conNameEdit.Text()
								},
							},
							TextLabel{Text: ".mp4"},
						},
					},
					PushButton{
						Text: "Concat these",
						OnClicked: func() {
							if len(mw.vidList) < 2 {
								walk.MsgBox(mw, "Error", "Select at least 2 videos", walk.MsgBoxIconInformation)
							} else {
								go mw.concatVideo(mw.conVidName)
							}
						},
					},
					PushButton{
						Text: "Open file location",
						OnClicked: func() {
							openFileLocation(mw.exPath, mw.conVidName)
						},
					},
				},
			},
			TextLabel{
				//ColumnSpan: 1,

				AssignTo: &mw.progressBarLabel,
				Text:     "nVidEditor",
			}, /*
				Composite{
					AssignTo: &mw.progressComposite,
					Layout: Grid{
						MarginsZero: true,
						SpacingZero: true,
						Columns:     1,
					},
					Children: []Widget{

						CustomWidget{
							ColumnSpan:          1,
							AssignTo:            &mw.progressBar,
							MaxSize:             Size{Height: 20},
							ClearsBackground:    true,
							InvalidatesOnResize: true,
							PaintPixels:         mw.paintProgressbar,
						},
					},
				},
			*/
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	if fileExists("./resources/ico256.ico") {
		icon, _ := walk.NewIconFromFile("./resources/ico256.ico")
		mw.SetIcon(icon)
	}

	mw.initNotifyIcon()
	defer mw.ni.Dispose()

	ffmpegPath = "ffmpeg.exe"
	ffprobePath = "ffprobe.exe"
	mw.askAboutFfmpeg()

	mw.Run()
}

/*











 CUTTING FUNCS
*/

func (mw *MyMainWindow) cutOpenFile() error {
	dlg := new(walk.FileDialog)
	dlg.FilePath = mw.cutPrevFilePath
	//dlg.Filter = "Image Files (*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff)|*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff"
	dlg.Filter = "Image Files (*.mp4;*.mkv;*.wmv)|*.mp4;*.mkv;*.wmv"
	dlg.Title = "Select a Video"
	if ok, err := dlg.ShowOpen(mw); err != nil {
		return err
	} else if !ok {
		return nil
	}
	go mw.cutSetVideo(dlg.FilePath)
	return nil
}

func (mw *MyMainWindow) cutSetVideo(filepath string) {
	mw.cutPrevFilePath = filepath
	mw.cutTextLabel.SetText(filepath)
	fs := strings.SplitAfter(filepath, `\`)
	fn := fs[len(fs)-1]
	filename := fn[:len(fn)-4]
	mw.cutNameEdit.SetText(filename)
	//mw.vidLength.SetText(getVideoDuration(filepath))
	ivm, err := getThumbnail(filepath)
	if err != nil {
		log.Println(err)
	}
	go mw.addVideoToEditCut(filename, filepath, ivm)
	//removeThumbnail()
}

func (mw *MyMainWindow) addVideoToEditCut(filename, filepath string, ivm walk.Image) {
	mw.cutThumbIV.SetImage(ivm)
	mw.vidLength.SetText(getVideoDuration(filepath))
}

// FHHHUuuuhHHUH arguments are formatted automatically probably. Took me a while to understand that
func (mw *MyMainWindow) cutVideo(exPath, item, name, sh, sm, ss, eh, em, es string, ni walk.NotifyIcon) {
	fmt.Println("Starting the cutting")
	if fileExists(exPath + `\` + name + ".mp4") {
		timenow := time.Now()
		name = name + "-" + timeFormat(strconv.Itoa(timenow.Hour()), strconv.Itoa(timenow.Minute()), strconv.Itoa(timenow.Second()))
	}

	mw.progressCurrentLength = strconv.Itoa(convertToMM(eh, em, es) - convertToMM(sh, sm, ss))
	/*
		mw.progressFullLength = getVideoDurationMM(item)
		lengthInInt, _ := strconv.Atoi(mw.progressFullLength)
		if convertToMM(eh, em, es)+convertToMM(sh, sm, ss) > lengthInInt {
			mw.progressCurrentLength = mw.progressFullLength
		}
	*/
	start := timeFormat(sh, sm, ss)
	end := timeFormat(eh, em, es)

	//https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument

	/*
		args := []string{
			"/C",
			`ffmpeg.exe`,
			"-progress",
			"-",
			"-nostats",
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
			"2>&1",
		}

		cmd := exec.Command(`cmd`, args...)
	*/
	cmd := exec.Command(`cmd`)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		CmdLine:    `/C ` + ffmpegPath + ` -progress - -nostats -ss ` + formatStartTime(start) + ` -i "` + item + `" -to ` + formatEndTime(start, end) + ` -c copy -f mp4 "` + name + `.mp4" 2>&1`,
	}
	fmt.Println("cmd.sysprocattr + + + ", cmd.SysProcAttr.CmdLine)
	cmd.Dir = exPath
	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.CmdLine = `/C ffmpeg.exe -progress - -nostats -ss ` + formatStartTime(start) + ` -i "` + item + `" -to ` + formatEndTime(start, end) + ` -c copy -f mp4 "` + name + `.mp4" 2>&1`
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		var stdin bytes.Buffer
		cmd.Stdin = &stdin
	*/
	pipe, _ := cmd.StdoutPipe()

	fmt.Println("command executing: ", cmd)
	//fmt.Println("command executing: ", cmd.SysProcAttr.CmdLine)
	/*
		// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
		var stdoutBuf bytes.Buffer
		// var stderrBuf bytes.Buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		// cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	*/
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
	}

	// Getting Progress from ffmpeg stdout
	fmt.Println("READER READING")
	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		if strings.Contains(line, "out_time_ms") {
			reg, _ := regexp.Compile("[^0-9]+")
			currentMM := reg.ReplaceAllString(line, "")
			proge := getProgress(currentMM, mw.progressCurrentLength)
			//	mw.progressBar.SetWidthPixels(proge * mw.progressComposite.BoundsPixels().Width / 100)
			mw.progressBarLabel.SetText(strconv.Itoa(proge))
		}
		line, err = reader.ReadString('\n')
	}

	err = cmd.Wait() // Change the system for waiting. This freezes the whole application
	if err != nil {
		fmt.Println("error after wait", err)
	}

	mw.progressBarLabel.SetText("DONE")
	//mw.progressBar.SetWidthPixels(mw.progressComposite.ClientBounds().Width)
	/*
		outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
		fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	*/
	// fmt.Println(cmd.Stdout)
	if err := ni.ShowInfo("Video cutting done!", "Thanks for using Nve!"); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(stderr)
	// https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
	// fmt.Printf("%s", cmd.SysProcAttr.CmdLine)
	fmt.Printf("%s", cmd)
}

/*











// Some videos can't be the first. Maybe handling them first works


CONCATTING FUNCS
*/
func (mw *MyMainWindow) concatOpenFile() error {
	dlg := new(walk.FileDialog)
	dlg.FilePath = mw.conPrevFilePath
	//dlg.Filter = "Image Files (*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff)|*.emf;*.bmp;*.exif;*.gif;*.jpeg;*.jpg;*.png;*.tiff"
	dlg.Filter = "Image Files (*.mp4;*.mkv)|*.mp4;*.mkv"
	dlg.Title = "Select a Video"

	if ok, err := dlg.ShowOpen(mw); err != nil {
		return err
	} else if !ok {
		return nil
	}
	mw.concatSetVideo(dlg.FilePath)
	return nil
}

func (mw *MyMainWindow) concatSetVideo(filepath string) {
	fmt.Println("opened openVideoCon", filepath)
	mw.conPrevFilePath = filepath
	err := mw.conTextEdit.SetText(filepath)
	if err != nil {
		fmt.Println(err)
	}
	fs := strings.SplitAfter(filepath, `\`)
	fn := fs[len(fs)-1]
	filename := fn[:len(fn)-4]
	mw.conNameEdit.SetText(filename)
	//mw.vidLength.SetText(getVideoDuration(filepath))
	ivm, err := getThumbnail(filepath)
	if err != nil {
		log.Println(err)
	}
	mw.addVideoToList(filename, filepath, ivm)
}

type Com struct {
	*walk.Composite
}

func (mw *MyMainWindow) addVideoToList(filename, path string, img walk.Image) {
	com := new(Com)
	if err := (Composite{
		AssignTo: &com.Composite,
		Name:     filename,
		Border:   true,
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Label{
				TextAlignment: AlignCenter,
				Text:          filename,
			},
			ImageView{
				Image: img,
				Mode:  ImageViewModeCenter,
			},
			PushButton{
				Text: "Remove",
				OnClicked: func() {
					//fmt.Println(mw.vidListComp.Children().Index(com))
					ic := mw.vidListComp.Children().Index(com)
					mw.vidList = append(mw.vidList[:ic], mw.vidList[ic+1:]...)
					com.Dispose()
				},
			},
		},
	}).Create(NewBuilder(mw.vidListComp)); err != nil {
		log.Println(err)
	}
	mw.vidList = append(mw.vidList, path)
	fmt.Println(mw.vidList)
}

func (mw *MyMainWindow) concatVideo(name string) {
	mw.conWorking = true
	go mw.conWorkProgress()
	if fileExists(mw.exPath + `\` + name + ".mp4") {
		timenow := time.Now()
		name = name + "-" + timeFormat(strconv.Itoa(timenow.Hour()), strconv.Itoa(timenow.Minute()), strconv.Itoa(timenow.Second()))
	}
	var intersToConcat string
	for i, n := range mw.vidList {
		fmt.Println("Generating intermediate file ", i, "...")
		args := []string{
			"/C",
			ffmpegPath,
			"-i",
			n,
			"-c",
			"copy",
			"-bsf:v",
			"h264_mp4toannexb",
			"-f",
			"mpegts",
			"intermediate" + strconv.Itoa(i) + ".ts",
		}
		cmd := exec.Command(`cmd`, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
		intersToConcat = intersToConcat + "intermediate" + strconv.Itoa(i) + ".ts|"
	}
	intersToConcat = intersToConcat[:len(intersToConcat)-1]
	fmt.Println("Starting concatinating files...")
	/*
		foreach ($d in $myarray) {
			$i++
			Write-Host('TRIED TO RUN WITH :: ' + $d)
			ffmpeg -i $d -c copy -bsf:v h264_mp4toannexb -f mpegts intermediate$i.ts
			$teksti = $teksti + "intermediate" + $i + ".ts|"
			# $teksti = $teksti + $d + '|'
			}
			ffmpeg -i "concat:$teksti" -c copy -bsf:a aac_adtstoasc $op
	*/
	//https://stackoverflow.com/questions/28954729/exec-with-double-quoted-argument
	cmd := exec.Command(`cmd`)
	cmd.Dir = mw.exPath
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmd.SysProcAttr.CmdLine = `/C ` + ffmpegPath + ` -i "concat:` + intersToConcat + `" -c copy -bsf:a aac_adtstoasc ` + name + `.mp4`
	//cmd.SysProcAttr.CmdLine = `/C ffmpeg.exe -i "concat:` + intersToConcat + `" -g 120 -keyint_min 4 -c copy -bsf:a aac_adtstoasc ` + name + `.mp4`
	err := cmd.Start()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
	// Now that the icon is visible, we can bring up an info balloon.
	if err := mw.ni.ShowInfo("Video concatting done!", "Thanks for using Nve!"); err != nil {
		log.Fatal(err)
	}
	mw.conWorking = false
	// https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
	//fmt.Printf("%s", cmd)

	removeIntermediates()
}

func (mw *MyMainWindow) conWorkProgress() {
	for mw.conWorking == true {
		mw.progressBarLabel.SetText("working.")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
		mw.progressBarLabel.SetText("working..")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
		mw.progressBarLabel.SetText("working..")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
		mw.progressBarLabel.SetText("working...")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
		mw.progressBarLabel.SetText("working....")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
		mw.progressBarLabel.SetText("working.....")
		if mw.conWorking == false {
			break
		}
		time.Sleep(1 * time.Second)
	}
	mw.progressBarLabel.SetText("DONE")
} // Errors won't stop this maybe.

/*











NEXT -> OPEN SIDE BAR WITH DIRECTORY BAR TO SET UP DESTINATION FOLDER FOR OUTPUT

OTHER FUNCS
*/

func (mw *MyMainWindow) aboutAction_Triggered() {
	walk.MsgBox(mw,
		"nVidEditor",
		"by Tuomo Miettinen - miettinen.codes",
		walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *MyMainWindow) helpAction_Triggered() {
	walk.MsgBox(mw,
		"nVidEditor",
		"Install ffmpeg from https://www.ffmpeg.org/download.html\r\nOr you can drop ffmpeg.exe and ffprobe.exe into ./resources/ffmpeg -named folder",
		walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *MyMainWindow) askAboutFfmpeg() {
	if !isFfmpegInstalled() {
		if !fileExists(`resources\ffmpeg\ffmpeg.exe`) || !fileExists(`resources\ffmpeg\ffprobe.exe`) {
			switch walk.MsgBox(mw, "Hey!", "You don't seem to have ffmpeg  installed.\nPut ffmpeg.exe and ffprobe.exe into resources/ffmpeg folder to continue.\nOr install ffmpeg to your computer.", walk.MsgBoxYesNo) {
			case walk.DlgCmdYes:
				fmt.Println("Yes Moi")
			case walk.DlgCmdNo:
				fmt.Println("No Moi")
			}
		}
		//walk.MsgBox(mw, "Error", "You don't seem to have ffmpeg  installed. \n\r Place ffmpeg.exe and ffprobe.exe into ffmpeg folder next to nVideEditor.exe", walk.MsgBoxIconInformation)
		/*
			switch walk.MsgBox(
				mw,
				"Hey!",
				"You don't seem to have ffmpeg  installed. \n\rPlace ffmpeg.exe and ffprobe.exe into ffmpeg folder next to nVideEditor.exe\n\rDo you want to create a ffmpeg folder?",
				walk.MsgBoxYesNoCancel,
			) {
			case walk.DlgCmdYes:
				fmt.Println("moi1")
				createFfmpegFolder()
			case walk.DlgCmdNo:
				fmt.Println("moi2")
			case walk.DlgCmdCancel:
				fmt.Println("moi3")
			}
		*/

		/*
			switch walk.MsgBox(mw, "Hey!", "You don't seem to have ffmpeg  installed. \n\rYou need to have ffmpeg and ffprobe working \n\rDo you want to create a ffmpeg folder?", walk.MsgBoxYesNo) {
			case walk.DlgCmdYes:
				createFfmpegFolder()
			case walk.DlgCmdNo:
				fmt.Println("moi2")
			}
		*/
		ffmpegPath = `resources\ffmpeg\ffmpeg.exe`
		ffprobePath = `resources\ffmpeg\ffprobe.exe`
	}
}

func (mw *MyMainWindow) initNotifyIcon() {
	// Create the notify icon and make sure we clean it up on exit.
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	// We load our icon from a file.
	if fileExists("./resources/ico256.ico") {
		icon, err := walk.Resources.Icon("./resources/ico256.ico")
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
		}
		openFileLocation(mw.exPath, mw.vidName)
	})
	openLocationAction := walk.NewAction()
	if err := openLocationAction.SetText("&Open File Location"); err != nil {
		log.Fatal(err)
	}
	openLocationAction.Triggered().Attach(func() { openFileLocation(mw.exPath, mw.vidName) })
	if err := mw.ni.ContextMenu().Actions().Add(openLocationAction); err != nil {
		log.Fatal(err)
	}
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
