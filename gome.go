/*
Package gome provides a minimal and simple library for setting up a
graphical application using OpenGL. All functions should be called on the main
OS thread. Init locks the goroutine to the main OS thread, so calling that early
in main ensures that any subsequent calls in main are on the right thread.

The main loop of the application then looks like this:

    if err := gome.Init(); err != nil {
        // handle error
    }
    defer gome.Terminate()

    initGL()
    render()
    gome.Tick()

    gome.Window.Show()

    for gome.Tick() {
        update()
        render()
    }
    if err := gome.GetError(); err != nil {
        // handle error
    }

A time.Ticker can be used to limit framerate.
*/
package gome

import (
    "errors"
    "github.com/go-gl/gl"
    "github.com/go-gl/glfw3"
    "github.com/go-gl/glu"
    "runtime"
)

var (
    ErrGLFW3Initialize = errors.New("could not initialise GLFW3")
    ErrGLEWInitialize  = errors.New("could not initialise GLEW")
)

type glError gl.GLenum

func (e glError) Error() string {
    // it seems like GLU cannot be built under Go 1.3 (had to patch it)
    m, err := glu.ErrorString(gl.GLenum(e))
    if err != nil {
        return err.Error()
    }
    return m
}

var tickError error

// GetError polls OpenGL for an error and returns that.
func GetError() error {
    if e := tickError; e != nil {
        tickError = nil
        return e
    }
    if code := gl.GetError(); code != 0 {
        return glError(code)
    }
    return nil
}

// Window is the main window of the application. This is created automatically
// by Init and has dimensions 800x600 by default. It is hidden by default, so
// the application should call gome.Window.Show() after any initialisation code.
var Window *glfw3.Window

// ShouldClose reflects whether the main loop should end. Setting ShouldClose to
// true causes gome.Tick to return false, which should end the main loop.
var ShouldClose = false

// Init initialises GLFW3 and OpenGL and creates the main window (see Window).
// After this has returned OpenGL functions as well as gome.Tick can be used.
// It also locks the current OS thread (see runtime.LockOSThread).
func Init() error {
    runtime.LockOSThread()

    if !glfw3.Init() {
        return ErrGLFW3Initialize
    }

    // request OpenGL 3.2 (forward compatible, core)
    glfw3.WindowHint(glfw3.ContextVersionMajor, 3)
    glfw3.WindowHint(glfw3.ContextVersionMinor, 2) // or 3
    glfw3.WindowHint(glfw3.OpenglForwardCompatible, 1)
    glfw3.WindowHint(glfw3.OpenglProfile, glfw3.OpenglCoreProfile)

    // glfw3.WindowHint(glfw3.Visible, 0)
    window, err := glfw3.CreateWindow(800, 600, "Gome", nil, nil)
    if err != nil {
        return err
    }
    window.MakeContextCurrent()
    Window = window

    glfw3.SwapInterval(1)

    if err := gl.Init(); err != 0 {
        return ErrGLEWInitialize
    }

    errcode := gl.GetError()
    for errcode == gl.INVALID_ENUM {
        errcode = gl.GetError()
    }
    if errcode != 0 {
        return glError(errcode)
    }
    return nil
}

// Tick swaps the buffers of the main window and polls GLFW3 for events. It
// returns true if the main loop should continue and false otherwise. It only
// returns false if ShouldClose is true, the window is being closed or if
// OpenGL reports an error.
func Tick() bool {
    if err := GetError(); err != nil {
        tickError = err
        return false
    }
    if ShouldClose || Window.ShouldClose() {
        return false
    }
    Window.SwapBuffers()
    glfw3.PollEvents()
    return true
}

// Terminate cleans up and terminates GLFW3. It should be called after the main
// loop has finished, e.g. by deferring it in the main function.
func Terminate() {
    Window.Destroy()
    glfw3.Terminate()
}
