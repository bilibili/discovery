package registry

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/model"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/ecode"

	. "github.com/smartystreets/goconvey/convey"
)

var reg = &model.ArgRegister{AppID: "main.arch.test", Hostname: "reg", Zone: "sh0001", Env: "pre", Status: 1}
var regH1 = &model.ArgRegister{AppID: "main.arch.test", Hostname: "regH1", Zone: "sh0001", Env: "pre", Status: 1}

var reg2 = &model.ArgRegister{AppID: "main.arch.test2", Hostname: "reg2", Zone: "sh0001", Env: "pre", Status: 1}

var arg = &model.ArgRenew{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Hostname: "reg"}
var cancel = &model.ArgCancel{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Hostname: "reg"}
var cancel2 = &model.ArgCancel{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Hostname: "regH1"}

func TestMain(m *testing.M) {
	flag.Set("conf", "./")
	flag.Parse()
	paladin.Init()
	m.Run()
	os.Exit(0)
}
func TestReigster(t *testing.T) {
	i := model.NewInstance(reg)
	register(t, i)
}

func TestDiscovery(t *testing.T) {
	i1 := model.NewInstance(reg)
	i2 := model.NewInstance(regH1)
	r := register(t, i1, i2)
	Convey("test discovery", t, func() {
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", AppID: []string{"main.arch.test"}, Hostname: "test"}
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		info, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		So(len(info.Instances["sh0001"]), ShouldEqual, 2)
		ch, _, _, err := r.Polls(pollArg)
		So(err, ShouldBeNil)
		apps := <-ch
		So(len(apps["main.arch.test"].Instances["sh0001"]), ShouldEqual, 2)
		pollArg.LatestTimestamp[0] = apps["main.arch.test"].LatestTimestamp
		fmt.Println(apps["main.arch.test"])
		r.Cancel(cancel)
		ch, _, _, err = r.Polls(pollArg)
		So(err, ShouldBeNil)
		apps = <-ch
		So(len(apps["main.arch.test"].Instances), ShouldEqual, 1)
		pollArg.LatestTimestamp[0] = apps["main.arch.test"].LatestTimestamp
		r.Cancel(cancel2)
	})
}

func TestRenew(t *testing.T) {
	src := model.NewInstance(reg)
	r := register(t, src)
	Convey("test renew", t, func() {
		i, ok := r.Renew(arg)
		So(ok, ShouldBeTrue)
		So(i, ShouldResemble, src)
	})
}

func BenchmarkRenew(b *testing.B) {
	var (
		i  *model.Instance
		ok bool
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, src := benchRegister(b)
			if i, ok = r.Renew(arg); !ok {
				b.Errorf("Renew(%v)", src.AppID)
			}
			benchCompareInstance(b, src, i)
		}
	})
}

func TestCancel(t *testing.T) {
	src := model.NewInstance(reg)
	r := register(t, src)
	Convey("test cancel", t, func() {
		i, ok := r.Cancel(cancel)
		So(ok, ShouldBeTrue)
		So(i, ShouldResemble, src)
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		_, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldResemble, ecode.NothingFound)
	})
}

func BenchmarkCancel(b *testing.B) {
	var (
		i   *model.Instance
		ok  bool
		err error
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, src := benchRegister(b)
			if i, ok = r.Cancel(cancel); !ok {
				b.Errorf("Cancel(%v) error", src.AppID)
			}
			benchCompareInstance(b, src, i)
			fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
			if _, err = r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status); err != ecode.NothingFound {
				b.Errorf("Fetch(%v) error(%v)", src.AppID, err)
			}
		}
	})
}

func TestFetchAll(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	Convey("test fetch all", t, func() {
		am := r.FetchAll()
		So(len(am), ShouldResemble, 1)
	})
}

func BenchmarkFetchAll(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := benchRegister(b)
			if am := r.FetchAll(); len(am) != 1 {
				b.Errorf("FetchAll() error")
			}
		}
	})
}

func TestFetch(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	Convey("test fetch", t, func() {
		fetchArg2 := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 1}
		c, err := r.Fetch(fetchArg2.Zone, fetchArg2.Env, fetchArg2.AppID, 0, fetchArg2.Status)
		So(err, ShouldBeNil)
		So(len(c.Instances), ShouldResemble, 1)
	})
}

func BenchmarkFetch(b *testing.B) {
	var (
		err error
		c   *model.InstanceInfo
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ := benchRegister(b)
			fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 1}
			if c, err = r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status); err != nil {
				b.Errorf("Fetch(%v) error(%v)", arg.AppID, err)
			}
			fetchArg2 := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 2}
			if c, err = r.Fetch(fetchArg2.Zone, fetchArg2.Env, fetchArg2.AppID, 0, fetchArg2.Status); err != nil {
				b.Errorf("Fetch(%v) error(%v)", arg.AppID, err)
			}
			_ = c
		}
	})
}

func TestPoll(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	Convey("test poll", t, func() {
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", AppID: []string{"main.arch.test"}, Hostname: "csq"}
		ch, _, _, err := r.Polls(pollArg)
		So(err, ShouldBeNil)
		c := <-ch
		So(len(c[pollArg.AppID[0]].Instances), ShouldEqual, 1)
	})
}

func TestPolls(t *testing.T) {
	i1 := model.NewInstance(reg)
	i2 := model.NewInstance(reg2)
	r := register(t, i1, i2)
	Convey("test polls", t, func() {
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", LatestTimestamp: []int64{0, 0}, AppID: []string{"main.arch.test", "main.arch.test2"}, Hostname: "csq"}
		ch, new, _, err := r.Polls(pollArg)
		So(err, ShouldBeNil)
		So(new, ShouldBeTrue)
		c := <-ch
		So(len(c), ShouldResemble, 2)
	})
}

func TestPollsChan(t *testing.T) {
	i1 := model.NewInstance(reg)
	i2 := model.NewInstance(reg2)
	r := register(t, i1, i2)

	Convey("test polls parallel", t, func(c C) {
		var (
			wg       sync.WaitGroup
			ch1, ch2 chan map[string]*model.InstanceInfo
			new      bool
			err      error
		)
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", LatestTimestamp: []int64{time.Now().UnixNano(), time.Now().UnixNano()}, AppID: []string{"main.arch.test", "main.arch.test2"}, Hostname: "csq"}
		ch1, new, _, err = r.Polls(pollArg)
		c.So(err, ShouldEqual, ecode.NotModified)
		c.So(new, ShouldBeFalse)
		c.So(ch1, ShouldNotBeNil)
		ch2, new, _, err = r.Polls(pollArg)
		c.So(err, ShouldEqual, ecode.NotModified)
		c.So(new, ShouldBeFalse)
		c.So(ch2, ShouldNotBeNil)
		// wait group
		wg.Add(2)
		go func() {
			res := <-ch1
			c.So(len(res), ShouldResemble, 1)
			wg.Done()
		}()
		go func() {
			res := <-ch2
			c.So(len(res), ShouldResemble, 1)
			wg.Done()
		}()
		// re register when 1s later, make sure latest_timestamp changed
		time.Sleep(time.Second)
		h1 := model.NewInstance(regH1)
		_ = r.Register(h1, 0)
		// wait
		wg.Wait()
	})
}

func BenchmarkPoll(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var (
				err error
				ch  chan map[string]*model.InstanceInfo
				c   map[string]*model.InstanceInfo
			)
			r, _ := benchRegister(b)
			pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", AppID: []string{"main.arch.test"}, Hostname: "csq"}
			if ch, _, _, err = r.Polls(pollArg); err != nil {
				b.Errorf("Poll(%v) error(%v)", arg.AppID, err)
			}
			if c = <-ch; len(c[pollArg.AppID[0]].Instances) != 1 {
				b.Errorf("Poll(%v) lenth error", arg.AppID)
			}
		}
	})
}

func TestBroadcast(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	Convey("test poll push connection", t, func() {
		go func() {
			Convey("must poll ahead of time", t, func() {
				time.Sleep(time.Microsecond * 5)
				var arg2 = &model.ArgRegister{AppID: "main.arch.test", Hostname: "go", Zone: "sh0001", Env: "pre", Status: 1}
				m2 := model.NewInstance(arg2)
				err2 := r.Register(m2, 0)
				So(err2, ShouldBeNil)
			})
		}()
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", AppID: []string{"main.arch.test"}, LatestTimestamp: []int64{time.Now().UnixNano()}}
		ch, _, _, err := r.Polls(pollArg)
		So(err, ShouldResemble, ecode.NotModified)
		c := <-ch
		So(len(c[pollArg.AppID[0]].Instances["sh0001"]), ShouldResemble, 2)
		So(c[pollArg.AppID[0]].Instances, ShouldNotBeNil)
		So(len(c[pollArg.AppID[0]].Instances["sh0001"]), ShouldResemble, 2)
	})
}

func BenchmarkBroadcast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var (
			err  error
			err2 error
			ch   chan map[string]*model.InstanceInfo
			c    map[string]*model.InstanceInfo
		)
		r, _ := benchRegister(b)
		go func() {
			time.Sleep(time.Millisecond * 1)
			var arg2 = &model.ArgRegister{AppID: "main.arch.test", Hostname: "go", Zone: "sh0001", Env: "pre", Status: 1}
			m2 := model.NewInstance(arg2)
			if err2 = r.Register(m2, 0); err2 != nil {
				b.Errorf("Reigster(%v) error(%v)", m2.AppID, err2)
			}
		}()
		pollArg := &model.ArgPolls{Zone: "sh0001", Env: "pre", AppID: []string{"main.arch.test"}, LatestTimestamp: []int64{time.Now().UnixNano()}}
		if ch, _, _, err = r.Polls(pollArg); err != nil && err != ecode.NotModified {
			b.Errorf("Poll(%v) error(%v)", pollArg.AppID, err)
		}
		c = <-ch
		if len(c[pollArg.AppID[0]].Instances) != 2 {
			b.Errorf("Poll(%v) length error", pollArg.AppID)
		}
		if c[pollArg.AppID[0]].Instances == nil {
			b.Errorf("Poll(%v) zone instances nil error", pollArg.AppID)
		}
		if len(c[pollArg.AppID[0]].Instances["sh0001"]) != 2 {
			b.Errorf("Poll(%v) zone instances length error", pollArg.AppID)
		}
	}
}

func TestSet(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	changes := &model.ArgSet{Zone: "sh0001", Env: "pre", AppID: "main.arch.test"}
	changes.Hostname = []string{"reg"}
	changes.Status = []int64{1}
	Convey("test setstatus to 1", t, func() {
		ok := r.Set(changes)
		So(ok, ShouldBeTrue)
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		c, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		So(c.Instances["sh0001"][0].Status, ShouldResemble, uint32(1))
	})
	changes = &model.ArgSet{Zone: "sh0001", Env: "pre", AppID: "main.arch.test"}
	changes.Hostname = []string{"reg"}
	changes = &model.ArgSet{Zone: "sh0001", Env: "pre", AppID: "main.arch.test"}
	changes.Hostname = []string{"reg"}
	changes.Metadata = []string{`{"weight":"11"}`}
	Convey("test set metadata weight to 11", t, func() {
		ok := r.Set(changes)
		So(ok, ShouldBeTrue)
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		c, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		So(c.Instances["sh0001"][0].Metadata["weight"], ShouldResemble, "11")
	})
	i1 := model.NewInstance(regH1)
	_ = r.Register(i1, 0)
	changes = &model.ArgSet{Zone: "sh0001", Env: "pre", AppID: "main.arch.test"}
	changes.Hostname = []string{"reg", "regH1"}
	changes.Metadata = []string{`{"weight":"12"}`, `{"weight":"13"}`}
	Convey("test set multi instance's color to blue and metadata weight to 12", t, func() {
		ok := r.Set(changes)
		So(ok, ShouldBeTrue)
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		c, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		for _, ins := range c.Instances["sh0001"] {
			if ins.Hostname == "reg" {
				So(ins.Metadata["weight"], ShouldResemble, "12")
			} else if ins.Hostname == "regH1" {
				So(ins.Metadata["weight"], ShouldResemble, "13")
			}
		}
	})
}

func BenchmarkSet(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var (
				c   *model.InstanceInfo
				err error
				ok  bool
			)
			r, _ := benchRegister(b)
			changes := &model.ArgSet{Zone: "sh0001", Env: "pre", AppID: "main.arch.test"}
			changes.Hostname = []string{"reg"}
			changes.Status = []int64{1}

			if ok = r.Set(changes); !ok {
				b.Errorf("SetStatus(%v) error", arg.AppID)
			}
			fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
			if c, err = r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status); err != nil {
				b.Errorf("Fetch(%v) error(%v)", fetchArg.AppID, err)
				b.FailNow()
			}
			if c.Instances["sh0001"][0].Status != 1 {
				b.Errorf("SetStatus(%v) change error", fetchArg.AppID)
			}
		}
	})
}

func TestResetExp(t *testing.T) {
	i := model.NewInstance(reg)
	r := register(t, i)
	Convey("test ResetExp", t, func() {
		r.resetExp()
		So(r.gd.expPerMin, ShouldResemble, int64(2))
	})
}

func benchCompareInstance(b *testing.B, src *model.Instance, i *model.Instance) {
	if src.AppID != i.AppID || src.Env != i.Env || src.Hostname != i.Hostname {
		b.Errorf("instance compare error")
	}
}

func register(t *testing.T, is ...*model.Instance) (r *Registry) {
	Convey("test register", t, func() {
		r = NewRegistry(&conf.Config{})
		var num int
		for _, i := range is {
			err := r.Register(i, 0)
			So(err, ShouldBeNil)
			if i.AppID == "main.arch.test" {
				num++
			}
		}
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		instancesInfo, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		So(len(instancesInfo.Instances["sh0001"]), ShouldResemble, num)
	})
	return r
}

func benchRegister(b *testing.B) (r *Registry, i *model.Instance) {
	r = NewRegistry(&conf.Config{})
	i = model.NewInstance(reg)
	if err := r.Register(i, 0); err != nil {
		b.Errorf("Reigster(%v) error(%v)", i.AppID, err)
	}
	return r, i
}

func TestEvict(t *testing.T) {
	Convey("test evict for protect", t, func() {
		r := NewRegistry(&conf.Config{})
		m := model.NewInstance(reg)
		// promise the renewtime of instance is expire
		m.RenewTimestamp -= 100
		err := r.Register(m, 0)
		So(err, ShouldBeNil)
		// move up the statistics of heartbeat for evict
		r.gd.facLastMin = r.gd.facInMin
		r.evict()
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 3}
		c, err := r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldBeNil)
		// protect
		So(len(c.Instances), ShouldResemble, 1)
	})
}

func TestEvict2(t *testing.T) {
	Convey("test evict for cancel", t, func() {
		r := NewRegistry(&conf.Config{})
		m := model.NewInstance(reg)
		err := r.Register(m, 0)
		So(err, ShouldBeNil)
		_, ok := r.Renew(arg)
		So(ok, ShouldBeTrue)
		// promise the renewtime of instance is expire
		m.RenewTimestamp -= int64(time.Second * 100)
		_ = r.Register(m, 0)
		// move up the statistics of heartbeat for evict
		r.gd.facLastMin = r.gd.facInMin
		r.evict()
		fetchArg := &model.ArgFetch{Zone: "sh0001", Env: "pre", AppID: "main.arch.test", Status: 1}
		_, err = r.Fetch(fetchArg.Zone, fetchArg.Env, fetchArg.AppID, 0, fetchArg.Status)
		So(err, ShouldResemble, ecode.NothingFound)
	})
}
