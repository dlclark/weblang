////////////////////////////
//// helper types 

type todo struct {
    isCompleted bool
    isEditing   bool
    title       string
}

type activeFilter enum {
    None
    Active
    Completed
}

////////////////////////////
//// page level vars are the implicit model for the page

var todos []todo = []todo{}
var curFilter = activeFilter.None 

////////////////////////////
//// TODO: Do we need to have the explicitly defined template backing vars?
//// or can we figure it out automatically by pre-parsing the template?
//// By defining these expicitly we can remove a step in the compile and just process the code first 
//// then process the template based on the fully parsed code

var newTodo html.Input
var todoEditBox html.Input
var toggle-all html.Input


////////////////////////////
//// helper funcs

func activeTodoCount() int {
    return todos.Filter(t => !t.isCompleted).Length()
} 

func completedTodoCount() int {
    return todos.Filter(t => t.isCompleted).Length()
}

// return true to be included in the results
func toCurFilter(todo t) bool {
    switch curFilter {
        case Active:
            return !t.isCompleted 
        case Completed:
            return t.isCompleted
        default:
            return true
    }
}

////////////////////////////
//// events

func addTodo(e events.KeyUp) {
    todos = append(todos, todo{ 
        title: newTodo.Value.Trim(),
    })
    newTodo.value = ""
}

func allDone() {
    for _, t := range todos {
        t.isCompleted = toggleAll.Checked
    }
}

func removeTodo(t todo) {
    remove(todos, t)
}

func editTodo(t todo) {
    t.isEditing = true
    todoEditBox.Value = t.title
}

func doneEdit(t todo) {
    if !t.isEditing {
        return
    }

    t.isEditing = false 
    if val := todoEditBox.Value.Trim(); len(val) > 0 {
        t.title = val
    } else {
        removeTodo(t)
    }
}

func cancelEdit(t todo) {
    t.isEditing = false
}

func removeCompleted() {
    todos = todos.Filter(t => !t.isCompleted).ToSlice()
}