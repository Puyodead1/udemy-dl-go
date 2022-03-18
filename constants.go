package main

import "path/filepath"

// Udemy
const COURSE_URL = "https://{portal_name}.udemy.com/api-2.0/courses/{course_id}/cached-subscriber-curriculum-items?fields[asset]=results,title,external_url,time_estimation,download_urls,slide_urls,filename,asset_type,captions,media_license_token,course_is_drmed,media_sources,stream_urls,body&fields[chapter]=object_index,title,sort_order&fields[lecture]=id,title,object_index,asset,supplementary_assets,view_html&page_size=10000"
const COURSE_INFO_URL = "https://{portal_name}.udemy.com/api-2.0/courses/{course_id}/"
const COURSE_SEARCH_URL = "https://{portal_name}.udemy.com/api-2.0/users/me/subscribed-courses?fields[course]=id,url,title,published_title&page=1&page_size=500&search={course_name}"
const SUBSCRIBED_COURSES_URL = "https://{portal_name}.udemy.com/api-2.0/users/me/subscribed-courses/?ordering=-last_accessed&fields[course]=id,title,url&page=1&page_size=12"
const MY_COURSES_URL = "https://{portal_name}.udemy.com/api-2.0/users/me/subscribed-courses?fields[course]=id,url,title,published_title&ordering=-last_accessed,-access_time&page=1&page_size=10000"
const COLLECTION_URL = "https://{portal_name}.udemy.com/api-2.0/users/me/subscribed-courses-collections/?collection_has_courses=True&course_limit=20&fields[course]=last_accessed_time,title,published_title&fields[user_has_subscribed_courses_collection]=@all&page=1&page_size=1000"
const LOGIN_URL = "https://www.udemy.com/join/login-popup/?ref=&display_type=popup&loc"

// FFMPEG Windows
const FFMPEG_WIN_LATEST_VERSION_URL = "https://www.gyan.dev/ffmpeg/builds/git-version"
const FFMPEG_WIN_SHA_URL = "https://www.gyan.dev/ffmpeg/builds/packages/ffmpeg-%s-essentials_build.7z.sha256"
const FFMPEG_WIN_URL = "https://www.gyan.dev/ffmpeg/builds/packages/ffmpeg-%s-essentials_build.7z"

// FFMPEG Mac
const FFMPEG_MAC_INFO_URL = "https://evermeet.cx/ffmpeg/info/ffmpeg/snapshot"   // gets information about the latest snapshot
const FFMPEG_MAC_VERSION_INFO_URL = "https://evermeet.cx/ffmpeg/info/ffmpeg/%s" // gets information about a specific version

// Paths
var FFMPEG_BIN_DIRECTORY = filepath.Join("bin", "ffmpeg")
