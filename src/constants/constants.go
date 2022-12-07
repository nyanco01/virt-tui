package constants

const (
    Kilo    uint64 = 1000
    Mega    uint64 = 1000 * 1000
    Giga    uint64 = 1000 * 1000 * 1000
)

const (
    Kibi    uint64 = 1024
    Mebi    uint64 = 1024 * 1024
    Gibi    uint64 = 1024 * 1024 * 1024
)

const (
    LeftTop         string = "╔══"
    LeftDown        string = "╚══"
    RightTop        string = "══╗"
    RightDown       string = "══╝"
)

const (
    LeftTriangle    rune = 'ᐊ'
    RightTraiangle  rune = 'ᐅ'
    UpTraiangle     rune = 'ᐃ'
    DownTraiangle   rune = 'ᐁ'
)

const (
    CPUAffinity         string = "Requested operation is not valid: cpu affinity is not supported"
    AlreadyStarted      string = "Requested operation is not valid: domain is already running"
    DomainNotRunning    string = "Requested operation is not valid: domain is not running"
    UnableReadMonitor   string = "Unable to read from monitor: Connection reset by peer"
)
