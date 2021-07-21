package main

import (
	"fmt"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"
	"regexp"
	"strings"
	"github.com/TheTitanrain/w32"
)

func banner() {
	fmt.Printf("\n###### Ableton AlwaysHotKeys######\n\n")
	fmt.Printf("sends alphanumeric characters to Ableton,\neven when typing in another program\n\n")
	fmt.Printf("made by Gianluca Elia for Daniel Sousa\n\n")
	fmt.Printf("########################################\n\n")
}

func main() {
	banner()
	ableton := FindWindow("ableton")
	if(ableton == 0) {
		fmt.Printf("Error: can't find Ableton window... quitting")
		return
	}
	fmt.Printf("Running!\n")
	kl := NewKeylogger()
	for {
		key := kl.GetKey()
		if !key.Empty {
			if (IsAlphanumeric(key)) {
				if (!IsActiveWindow(ableton)) {
					w32.PostMessage(ableton, w32.WM_CHAR, uintptr(key.Keycode), 0)
				}
			}
			fmt.Printf("%c", key.Rune)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func IsAlphanumeric(key Key) bool {
	code := key.Keycode
	isAlpha := code > 64 && code < 91
	isNum := code > 47 && code < 58
	return isAlpha || isNum
}

var (
	moduser32 = syscall.NewLazyDLL("user32.dll")

	procGetKeyboardLayout     = moduser32.NewProc("GetKeyboardLayout")
	procGetKeyboardState      = moduser32.NewProc("GetKeyboardState")
	procToUnicodeEx           = moduser32.NewProc("ToUnicodeEx")
	procGetKeyboardLayoutList = moduser32.NewProc("GetKeyboardLayoutList")
	procMapVirtualKeyEx       = moduser32.NewProc("MapVirtualKeyExW")
	procGetKeyState           = moduser32.NewProc("GetKeyState")
	procEnumWindows           = moduser32.NewProc("EnumWindows")
	procGetWindowText         = moduser32.NewProc("GetWindowTextW")
        procGetWindow             = moduser32.NewProc("GetWindow")
        procGetForegroundWindow   = moduser32.NewProc("GetForegroundWindow")
)

func GetWindowText(h w32.HWND) string{
    buf := make([]uint16, 200)
    res, _, _ := procGetWindowText.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(200))
    if (res == 0) {
	return ""
    }
    text := syscall.UTF16ToString(buf)
    return text	
}

func GetChild(h w32.HWND) w32.HWND {
	child, _, _ := procGetWindow.Call(uintptr(h), uintptr(5))
	return w32.HWND(child)
}

func IsActiveWindow(target w32.HWND) bool {
	activeWindowPtr, _, _ := procGetForegroundWindow.Call()
	activeWindow := w32.HWND(activeWindowPtr)
	child := GetChild(activeWindow)
	// fmt.Printf("active: %d, child: %d target: %d\n", activeWindow,child, target)
	return activeWindow == target || child == target
}

func FindWindow(title string) w32.HWND{
	var hwnd w32.HWND
	title = strings.ToLower(title)
	cb := syscall.NewCallback(func(h w32.HWND, p uintptr) uintptr {
		thisTitle := GetWindowText(h)
		match, _:= regexp.MatchString(title, strings.ToLower(thisTitle))
		if(match) {
			fmt.Printf("Found window: %s\n", thisTitle)
			child := GetChild(h)
			if (child != 0) {
			  hwnd = child
			  return 0
			}		
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return hwnd
}

// NewKeylogger creates a new keylogger depending on
// the platform we are running on (currently only Windows
// is supported)
func NewKeylogger() Keylogger {
	kl := Keylogger{}

	return kl
}

// Keylogger represents the keylogger
type Keylogger struct {
	lastKey int
}

// Key is a single key entered by the user
type Key struct {
	Empty   bool
	Rune    rune
	Keycode int
}

// GetKey gets the current entered key by the user, if there is any
func (kl *Keylogger) GetKey() Key {
	activeKey := 0
	var keyState uint16

	for i := 0; i < 256; i++ {
		keyState = w32.GetAsyncKeyState(i)

		// Check if the most significant bit is set (key is down)
		// And check if the key is not a non-char key (except for space, 0x20)
		if keyState&(1<<15) != 0 && !(i < 0x2F && i != 0x20) && (i < 160 || i > 165) && (i < 91 || i > 93) {
			activeKey = i
			break
		}
	}

	if activeKey != 0 {
		if activeKey != kl.lastKey {
			kl.lastKey = activeKey
			return kl.ParseKeycode(activeKey, keyState)
		}
	} else {
		kl.lastKey = 0
	}

	return Key{Empty: true}
}

// ParseKeycode returns the correct Key struct for a key taking in account the current keyboard settings
// That struct contains the Rune for the key
func (kl Keylogger) ParseKeycode(keyCode int, keyState uint16) Key {
	key := Key{Empty: false, Keycode: keyCode}

	// Only one rune has to fit in
	outBuf := make([]uint16, 1)

	// Buffer to store the keyboard state in
	kbState := make([]uint8, 256)

	// Get keyboard layout for this process (0)
	kbLayout, _, _ := procGetKeyboardLayout.Call(uintptr(0))

	// Put all key modifier keys inside the kbState list
	if w32.GetAsyncKeyState(w32.VK_SHIFT)&(1<<15) != 0 {
		kbState[w32.VK_SHIFT] = 0xFF
	}

	capitalState, _, _ := procGetKeyState.Call(uintptr(w32.VK_CAPITAL))
	if capitalState != 0 {
		kbState[w32.VK_CAPITAL] = 0xFF
	}

	if w32.GetAsyncKeyState(w32.VK_CONTROL)&(1<<15) != 0 {
		kbState[w32.VK_CONTROL] = 0xFF
	}

	if w32.GetAsyncKeyState(w32.VK_MENU)&(1<<15) != 0 {
		kbState[w32.VK_MENU] = 0xFF
	}

	_, _, _ = procToUnicodeEx.Call(
		uintptr(keyCode),
		uintptr(0),
		uintptr(unsafe.Pointer(&kbState[0])),
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(1),
		uintptr(1),
		uintptr(kbLayout))

	key.Rune, _ = utf8.DecodeRuneInString(syscall.UTF16ToString(outBuf))

	return key
}
