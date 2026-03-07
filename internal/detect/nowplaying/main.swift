import Foundation

// Dynamically load MediaRemote.framework (private framework)
guard let bundle = CFBundleCreate(
    kCFAllocatorDefault,
    NSURL(fileURLWithPath: "/System/Library/PrivateFrameworks/MediaRemote.framework")
) else {
    exit(1)
}

// Type aliases for MediaRemote functions
typealias MRMediaRemoteGetNowPlayingInfoFunc = @convention(c) (DispatchQueue, @escaping ([String: Any]) -> Void) -> Void
typealias MRMediaRemoteGetNowPlayingClientFunc = @convention(c) (DispatchQueue, @escaping (AnyObject?) -> Void) -> Void
typealias MRNowPlayingClientGetBundleIdentifierFunc = @convention(c) (AnyObject) -> NSString?
typealias MRMediaRemoteGetNowPlayingApplicationIsPlayingFunc = @convention(c) (DispatchQueue, @escaping (Bool) -> Void) -> Void

// Load function pointers
guard let pInfo = CFBundleGetFunctionPointerForName(bundle, "MRMediaRemoteGetNowPlayingInfo" as CFString),
      let pClient = CFBundleGetFunctionPointerForName(bundle, "MRMediaRemoteGetNowPlayingClient" as CFString),
      let pBundleID = CFBundleGetFunctionPointerForName(bundle, "MRNowPlayingClientGetBundleIdentifier" as CFString),
      let pIsPlaying = CFBundleGetFunctionPointerForName(bundle, "MRMediaRemoteGetNowPlayingApplicationIsPlaying" as CFString)
else {
    exit(1)
}

let MRMediaRemoteGetNowPlayingInfo = unsafeBitCast(pInfo, to: MRMediaRemoteGetNowPlayingInfoFunc.self)
let MRMediaRemoteGetNowPlayingClient = unsafeBitCast(pClient, to: MRMediaRemoteGetNowPlayingClientFunc.self)
let MRNowPlayingClientGetBundleIdentifier = unsafeBitCast(pBundleID, to: MRNowPlayingClientGetBundleIdentifierFunc.self)
let MRMediaRemoteGetNowPlayingApplicationIsPlaying = unsafeBitCast(pIsPlaying, to: MRMediaRemoteGetNowPlayingApplicationIsPlayingFunc.self)

// Use a serial queue to avoid data races between callbacks
let callbackQueue = DispatchQueue(label: "nowplaying-helper")
let group = DispatchGroup()

var title = ""
var artist = ""
var album = ""
var sourceID = ""
var isPlaying = false
var hasPlaybackRate = false
var playbackRate: Double = 0

// 1. Get playing state
group.enter()
MRMediaRemoteGetNowPlayingApplicationIsPlaying(callbackQueue) { playing in
    isPlaying = playing
    group.leave()
}

// 2. Get client bundle ID
group.enter()
MRMediaRemoteGetNowPlayingClient(callbackQueue) { client in
    if let client = client, let bid = MRNowPlayingClientGetBundleIdentifier(client) {
        sourceID = bid as String
    }
    group.leave()
}

// 3. Get now playing info
group.enter()
MRMediaRemoteGetNowPlayingInfo(callbackQueue) { info in
    title = info["kMRMediaRemoteNowPlayingInfoTitle"] as? String ?? ""
    artist = info["kMRMediaRemoteNowPlayingInfoArtist"] as? String ?? ""
    album = info["kMRMediaRemoteNowPlayingInfoAlbum"] as? String ?? ""
    if let rate = info["kMRMediaRemoteNowPlayingInfoPlaybackRate"] as? Double {
        hasPlaybackRate = true
        playbackRate = rate
    }
    group.leave()
}

let timeout = group.wait(timeout: .now() + 2)
if timeout == .timedOut {
    exit(1)
}

// PlaybackRate 0 means paused — only override if the key was actually present
if hasPlaybackRate && playbackRate == 0 {
    isPlaying = false
}

// Output JSON
let result: [String: Any] = [
    "title": title,
    "artist": artist,
    "album": album,
    "source_id": sourceID,
    "is_playing": isPlaying,
]

if let data = try? JSONSerialization.data(withJSONObject: result),
   let json = String(data: data, encoding: .utf8) {
    print(json)
}
