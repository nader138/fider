package postgres_test

import (
	"testing"
	"time"

	"github.com/getfider/fider/app/models/query"

	"github.com/getfider/fider/app/models/cmd"

	"github.com/getfider/fider/app"
	. "github.com/getfider/fider/app/pkg/assert"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/errors"
)

func TestNotificationStorage_TotalCount(t *testing.T) {
	ctx := SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	q := &query.CountUnreadNotifications{}

	err := bus.Dispatch(ctx, q)
	Expect(err).IsNil()
	Expect(q.Result).Equals(0)

	err = bus.Dispatch(demoTenantCtx, q)
	Expect(err).IsNil()
	Expect(q.Result).Equals(0)
}

func TestNotificationStorage_Insert_Read_Count(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	posts.SetCurrentTenant(demoTenant)
	posts.SetCurrentUser(jonSnow)
	post, _ := posts.Add("Title", "Description")

	addNotification1 := &cmd.AddNewNotification{User: aryaStark, Title: "Hello World", Link: "http://www.google.com.br", PostID: post.ID}
	addNotification2 := &cmd.AddNewNotification{User: aryaStark, Title: "Hello World", Link: "", PostID: post.ID}
	err := bus.Dispatch(jonSnowCtx, addNotification1, addNotification2)
	Expect(err).IsNil()

	q := &query.CountUnreadNotifications{}
	err = bus.Dispatch(aryaStarkCtx, q)
	Expect(err).IsNil()
	Expect(q.Result).Equals(2)

	Expect(bus.Dispatch(aryaStarkCtx, &cmd.MarkNotificationAsRead{ID: addNotification1.Result.ID})).IsNil()
	err = bus.Dispatch(aryaStarkCtx, q)
	Expect(err).IsNil()
	Expect(q.Result).Equals(1)

	Expect(bus.Dispatch(aryaStarkCtx, &cmd.MarkNotificationAsRead{ID: addNotification2.Result.ID})).IsNil()
	err = bus.Dispatch(aryaStarkCtx, q)
	Expect(err).IsNil()
	Expect(q.Result).Equals(0)
}

func TestNotificationStorage_GetActiveNotifications(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	posts.SetCurrentTenant(demoTenant)
	posts.SetCurrentUser(jonSnow)
	post, _ := posts.Add("Title", "Description")

	addNotification1 := &cmd.AddNewNotification{User: aryaStark, Title: "Hello World", Link: "http://www.google.com.br", PostID: post.ID}
	addNotification2 := &cmd.AddNewNotification{User: aryaStark, Title: "Another thing happened", Link: "http://www.google.com.br", PostID: post.ID}
	err := bus.Dispatch(jonSnowCtx, addNotification1, addNotification2)
	Expect(err).IsNil()

	activeNotifications := &query.GetActiveNotifications{}
	err = bus.Dispatch(aryaStarkCtx, activeNotifications)
	Expect(err).IsNil()
	Expect(activeNotifications.Result).HasLen(2)

	bus.Dispatch(aryaStarkCtx, &cmd.MarkNotificationAsRead{ID: activeNotifications.Result[0].ID})
	bus.Dispatch(aryaStarkCtx, &cmd.MarkNotificationAsRead{ID: activeNotifications.Result[1].ID})

	trx.Execute("UPDATE notifications SET updated_at = $1 WHERE id = $2", time.Now().AddDate(0, 0, -31), activeNotifications.Result[0].ID)

	err = bus.Dispatch(aryaStarkCtx, activeNotifications)
	Expect(err).IsNil()
	Expect(activeNotifications.Result).HasLen(1)
	Expect(activeNotifications.Result[0].Read).IsTrue()
}

func TestNotificationStorage_ReadAll(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	posts.SetCurrentTenant(demoTenant)
	posts.SetCurrentUser(jonSnow)

	post, _ := posts.Add("Title", "Description")

	addNotification1 := &cmd.AddNewNotification{User: aryaStark, Title: "Hello World", Link: "http://www.google.com.br", PostID: post.ID}
	addNotification2 := &cmd.AddNewNotification{User: aryaStark, Title: "Another thing happened", Link: "http://www.google.com.br", PostID: post.ID}
	err := bus.Dispatch(jonSnowCtx, addNotification1, addNotification2)
	Expect(err).IsNil()

	activeNotifications := &query.GetActiveNotifications{}
	err = bus.Dispatch(aryaStarkCtx, activeNotifications)
	Expect(err).IsNil()
	Expect(activeNotifications.Result).HasLen(2)

	err = bus.Dispatch(aryaStarkCtx, &cmd.MarkAllNotificationsAsRead{})
	Expect(err).IsNil()

	err = bus.Dispatch(aryaStarkCtx, activeNotifications)
	Expect(err).IsNil()
	Expect(activeNotifications.Result).HasLen(2)
	Expect(activeNotifications.Result[0].Read).IsTrue()
	Expect(activeNotifications.Result[1].Read).IsTrue()
}

func TestNotificationStorage_GetNotificationByID(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	posts.SetCurrentTenant(demoTenant)
	posts.SetCurrentUser(jonSnow)
	post, _ := posts.Add("Title", "Description")

	addNotification := &cmd.AddNewNotification{User: aryaStark, Title: "Hello World", Link: "http://www.google.com.br", PostID: post.ID}
	err := bus.Dispatch(jonSnowCtx, addNotification)
	Expect(err).IsNil()

	q := &query.GetNotificationByID{ID: addNotification.Result.ID}
	err = bus.Dispatch(aryaStarkCtx, q)
	Expect(err).IsNil()
	Expect(q.Result.Title).Equals("Hello World")
	Expect(q.Result.Link).Equals("http://www.google.com.br")
	Expect(q.Result.Read).IsFalse()
}

func TestNotificationStorage_GetNotificationByID_OtherUser(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	posts.SetCurrentTenant(demoTenant)
	posts.SetCurrentUser(jonSnow)
	post, _ := posts.Add("Title", "Description")

	addNotification := &cmd.AddNewNotification{User: jonSnow, Title: "Hello World", Link: "http://www.google.com.br", PostID: post.ID}
	err := bus.Dispatch(aryaStarkCtx, addNotification)
	Expect(err).IsNil()

	q := &query.GetNotificationByID{ID: addNotification.Result.ID}
	err = bus.Dispatch(aryaStarkCtx, q)
	Expect(errors.Cause(err)).Equals(app.ErrNotFound)
	Expect(q.Result).IsNil()
}
