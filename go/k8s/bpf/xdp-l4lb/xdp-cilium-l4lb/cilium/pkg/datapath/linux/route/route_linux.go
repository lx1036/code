//go:build linux

package route

import (
    "fmt"
    "github.com/vishvananda/netlink"
    "time"
)

const (
    // RouteReplaceMaxTries is the number of attempts the route will be
    // attempted to be added or updated in case the kernel returns an error
    RouteReplaceMaxTries = 10

    // RouteReplaceRetryInterval is the interval in which
    // RouteReplaceMaxTries attempts are attempted
    RouteReplaceRetryInterval = 100 * time.Millisecond
)

// Upsert adds or updates a Linux kernel route. The route described can be in
// the following two forms:
//
// direct:
//   prefix dev foo
//
// nexthop:
//   prefix via nexthop dev foo
//
// If a nexthop route is specified, this function will check whether a direct
// route to the nexthop exists and add if required. This means that the
// following two routes will exist afterwards:
//
//   nexthop dev foo
//   prefix via nexthop dev foo
//
// Due to a bug in the Linux kernel, the prefix route is attempted to be
// updated RouteReplaceMaxTries with an interval of RouteReplaceRetryInterval.
// This is a workaround for a race condition in which the direct route to the
// nexthop is not available immediately and the prefix route can fail with
// EINVAL if the Netlink calls are issued in short order.
//
// An error is returned if the route can not be added or updated.
func Upsert(route Route) error {
    var nexthopRouteCreated bool

    link, err := netlink.LinkByName(route.Device)
    if err != nil {
        return fmt.Errorf("unable to lookup interface %s: %s", route.Device, err)
    }

    routerNet := route.getNexthopAsIPNet()
    if routerNet != nil {
        if _, err := replaceNexthopRoute(route, link, routerNet); err != nil {
            return fmt.Errorf("unable to add nexthop route: %s", err)
        }

        nexthopRouteCreated = true
    }

    routeSpec := route.getNetlinkRoute()
    routeSpec.LinkIndex = link.Attrs().Index

    err = fmt.Errorf("routeReplace not called yet")

    // Workaround: See description of this function
    for i := 0; err != nil && i < RouteReplaceMaxTries; i++ {
        err = netlink.RouteReplace(&routeSpec)
        if err == nil {
            break
        }
        time.Sleep(RouteReplaceRetryInterval)
    }

    if err != nil {
        if nexthopRouteCreated {
            deleteNexthopRoute(route, link, routerNet)
        }
        return err
    }

    return nil
}
