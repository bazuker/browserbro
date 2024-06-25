package manager

import (
	"fmt"
	"strconv"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type browserConnector struct {
	browser                  *rod.Browser
	serverID                 int
	serviceURL               string
	userDataDir              string
	browserMonitoringEnabled bool
}

func newBrowserConnector(
	browser *rod.Browser,
	serverID int,
	serviceURL, userDataDir string,
	browserMonitoringEnabled bool,
) *browserConnector {
	return &browserConnector{
		browser:                  browser,
		serverID:                 serverID,
		serviceURL:               serviceURL,
		userDataDir:              userDataDir,
		browserMonitoringEnabled: browserMonitoringEnabled,
	}
}

func (br *browserConnector) Connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to launch browser: %v", r)
		}
	}()

	l, err := newManagedLauncher(
		br.serverID,
		br.serviceURL,
		br.userDataDir,
	)
	if err != nil {
		return err
	}
	l.NoSandbox(true)
	l.Set("disable-web-security")
	l.Set("disable-blink-features", "AutomationControlled")
	l.Delete("enable-automation")
	l.Delete("disable-site-isolation-trials")

	br.browser.Client(l.MustClient())
	err = br.browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	br.browser.MustIncognito()

	if br.browserMonitoringEnabled {
		launcher.Open(br.browser.ServeMonitor(":8889"))
	}

	return err
}

func newManagedLauncher(
	serverID int,
	serviceURL, userDataDir string,
) (l *launcher.Launcher, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to run managed browser: %v", r)
		}
	}()
	l = launcher.MustNewManaged(serviceURL).
		UserDataDir(userDataDir).
		Headless(false).
		Devtools(false).
		Leakless(true).XVFB("--server-num="+strconv.Itoa(serverID), "--server-args=-screen 0 1600x900x16")

	return l, err
}
