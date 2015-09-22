// -*- mode: go; coding: utf-8; tab-width: 4; -*-
package main

import (
	"encoding/json"
	"net/http"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	Id int
	Progress int
	mutex *sync.Mutex
}

func NewTask(id int) *Task {
	m := new(sync.Mutex)
	return &Task{id, 0, m}
}

func (t *Task) DoSomething() {
	if !t.NotDone() {
		// task ends
		return
	}

	time.Sleep(1 * time.Second)

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Progress++
}

func (t *Task) GetProgress() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Progress
}

func (t *Task) NotDone() bool {
	return t.Progress < 10
}

func TaskWorker(t *Task) {
	for t.NotDone() {
		t.DoSomething()
	}
}

type TaskController struct {
	AccessCounter int32
	TaskList []*Task
	m *sync.Mutex
}

func NewTaskController() *TaskController {
	return &TaskController{
		0,
		make([]*Task, 0, 32),
		new(sync.Mutex),
	}
}

func (t *TaskController) GenID() int {
	return int(atomic.AddInt32(&t.AccessCounter, 1))
}

func (t *TaskController) StartTask() int {
	id := t.GenID()
	task := NewTask(id)
	t.RegistTask(task)
	return id
}

func (t *TaskController) RegistTask(task *Task) {
	t.m.Lock()
	defer t.m.Unlock()
	t.TaskList = append(t.TaskList, task)
	go TaskWorker(task)
}

type TaskView struct {
	Id int `json:"id"`
	Progress int `json:"progress"`
	IsDone bool `json:"done"`
}

func (t *TaskController) GetTaskSummary() []TaskView {
	t.m.Lock()
	defer t.m.Unlock()
	view := make([]TaskView, 0, len(t.TaskList))
	for idx := range(t.TaskList) {
		ti := t.TaskList[idx]
		view = append(view, TaskView{ti.Id, ti.GetProgress(), !ti.NotDone()})
	}
	return view
}

type TaskServer struct {
	Controller *TaskController
}

func (s TaskServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s", r.RemoteAddr, r.URL.Path)
	switch {
	case strings.HasPrefix(r.URL.Path, "start"):
		s.Start(w, r)
	case strings.HasPrefix(r.URL.Path, "status"):
		s.Status(w, r)
	}
}

func (s TaskServer) Start(w http.ResponseWriter, r *http.Request) {
	id := s.Controller.StartTask()
	w.Header().Set("Content-Type", "application/json")
	jsonWriter := json.NewEncoder(w)
	err := jsonWriter.Encode(map[string]bool{
		"result": true,
	})
	if err != nil {
		log.Printf("json encode error: %v\n", err)
	}
	log.Printf("task started: id: %d\n", id)
}

func (s TaskServer) Status(w http.ResponseWriter, r *http.Request) {
	taskView := s.Controller.GetTaskSummary()
	w.Header().Set("Content-Type", "application/json")
	jsonWriter := json.NewEncoder(w)
	if err := jsonWriter.Encode(taskView); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json encode error: %v\n", err)
	}
	log.Printf("task: %d\n", len(taskView))
	log.Printf("id\tprogress\tDone\n")
	for idx := range(taskView) {
		t := taskView[idx]
		log.Printf("%d\t%d\t%v\n", t.Id, t.Progress, t.IsDone)
	}
}

func main() {
	controller := NewTaskController()
	http.Handle("/api/", http.StripPrefix("/api/", TaskServer{controller}))
	http.Handle("/",
		http.FileServer(http.Dir("static")))
	http.ListenAndServe(":8080", nil)
}
