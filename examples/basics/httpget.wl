package demo

imports (
    "http"
    "html/dom"
)

type Result union {
	Success []struct { 
        name string 
        years int
        rate float
         // if JsonKeys exists the json decoder will use these names so we can 
         // support keys that aren't valid idents (spaces, quotes, dashes, etc etc etc)
        JsonKeys map<string,string> { "name": "the-name" }
    }
	Err struct { errText string }
	Other string
}

type blah interface<T> {    // generics on interfaces
    Function<K>(in K) T     //functions on interfaces
    val T                   //fields on interfaces
}

type blahstruct struct {
    val int
}

func (b blahstruct) Function<K>(in K) int {
    return b.val
}

// set our default options for http requests
http.DefaultClient = { 
    RetryGetTransientErrCount: 10, 
    RetryGetStartDelay: time.Millisecond * 100,
    Timeout: time.Second * 10,
}

func SomeFunc(url string) {
    catch func(e error) {
        console.log(e)
    } 
    catch e => { console.log(e) }

    // http.Get() will retry on transient errors and throw if it fails to get a response from the server
    // after the specified timeout.
    // func (r http.Result) BodyAsJson<T>([rules]) T parses the body of the HTTP response as JSON 
    // using the rules passed in into a type T.
    rates := http.Get(url).RequireStatus(http.StatusCodes.OK).BodyAsJson<Result>(json.NoValidate)

    switch r := rates.(union) {
    case Success:
        var html string
        for _, rate := range r {
            html += `<tr><td>${rate.name}</td><td>${rate.years}</td><td>${rate.rate}%</td></tr>`
        }
        dom.MustGetElementById("rates").InnerHTML = html
    case Err:
        showTempErrorPanel(r.errText)
    case Other:
        showTempErrorPanel(r)
    }

    d := []blahstruct{
        {
            val: 1,
        },
    }
    Covariant(d)
}

func Covariant<T>(in []blah<T>) T {
    //do stuff with blah<T>
    return in[0].val
}

func (r Result) Test<T numeric>(in T) (T, T) {
    return in, in+1
}

type Queue struct<T> {
    items []T
}

func (q Queue<T>) Enqueue(item T) {
    q.items.Append(item)
}

func (q Queue<T>) Dequeue() Optional(T) {
    return q.items.Shift()
}

type StatusCodes enum int {
    Continue = 100
    OK = 200
    Created
    Accepted
    NotFound = 404
}

/* 
Multi line comments
yay!
*/
const SomeThing = 100
const SomeOtherThing = 1.1
const (
    Blah = "test"
    blah2 = 100
)


