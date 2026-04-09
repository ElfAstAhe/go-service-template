package transport

// application
const (
	MediaTypeApplicationAtomXML      string = "application/atom+xml"
	MediaTypeApplicationAtomCatXML   string = "application/atomcat+xml"
	MediaTypeApplicationEcmaScript   string = "application/ecmascript"
	MediaTypeApplicationEpubZip      string = "application/epub+zip"
	MediaTypeApplicationGZip         string = "application/gzip"
	MediaTypeApplicationJavaArchive  string = "application/java-archive"
	MediaTypeApplicationJavaScript   string = "application/javascript"
	MediaTypeApplicationJSON         string = "application/json"
	MediaTypeApplicationLdJSON       string = "application/ld+json"
	MediaTypeApplicationManifestJSON string = "application/manifest+json"
	MediaTypeApplicationMP4          string = "application/mp4"
	MediaTypeApplicationMSWord       string = "application/msword"
	MediaTypeApplicationOctetStream  string = "application/octet-stream"
	MediaTypeApplicationOgg          string = "application/ogg"
	MediaTypeApplicationPDF          string = "application/pdf"
	MediaTypeApplicationPkcs10       string = "application/pkcs10"
	MediaTypeApplicationPkcs7Mime    string = "application/pkcs7-mime"
	MediaTypeApplicationPkcs7Sig     string = "application/pkcs7-signature"
	MediaTypeApplicationPkcs8        string = "application/pkcs8"
	MediaTypeApplicationPostScript   string = "application/postscript"
	MediaTypeApplicationRdfXML       string = "application/rdf+xml"
	MediaTypeApplicationRssXML       string = "application/rss+xml"
	MediaTypeApplicationRTF          string = "application/rtf"
	MediaTypeApplicationSmilXML      string = "application/smil+xml"
	MediaTypeApplicationXhtmlXML     string = "application/xhtml+xml"
	MediaTypeApplicationXML          string = "application/xml"
	MediaTypeApplicationXmlDTD       string = "application/xml-dtd"
	MediaTypeApplicationXsltXML      string = "application/xslt+xml"
	MediaTypeApplicationZip          string = "application/zip"

	// Vendor specific
	MediaTypeApplicationVndAmazonEbook          string = "application/vnd.amazon.ebook"
	MediaTypeApplicationVndAppleInstallerXML    string = "application/vnd.apple.installer+xml"
	MediaTypeApplicationVndMozillaXulXML        string = "application/vnd.mozilla.xul+xml"
	MediaTypeApplicationVndMSExcel              string = "application/vnd.ms-excel"
	MediaTypeApplicationVndMSFontObject         string = "application/vnd.ms-fontobject"
	MediaTypeApplicationVndMSPowerpoint         string = "application/vnd.ms-powerpoint"
	MediaTypeApplicationVndOasisDocPresentation string = "application/vnd.oasis.opendocument.presentation"
	MediaTypeApplicationVndOasisDocSpreadsheet  string = "application/vnd.oasis.opendocument.spreadsheet"
	MediaTypeApplicationVndOasisDocText         string = "application/vnd.oasis.opendocument.text"

	// OpenXML
	MediaTypeApplicationVndOpenXMLDocPresentation string = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MediaTypeApplicationVndOpenXMLDocSpreadsheet  string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MediaTypeApplicationVndOpenXMLDocWord         string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	// X-types
	MediaTypeApplicationVndRar             string = "application/vnd.rar"
	MediaTypeApplicationVndVisio           string = "application/vnd.visio"
	MediaTypeApplicationX7zCompressed      string = "application/x-7z-compressed"
	MediaTypeApplicationXAbiword           string = "application/x-abiword"
	MediaTypeApplicationXBZip              string = "application/x-bzip"
	MediaTypeApplicationXBZip2             string = "application/x-bzip2"
	MediaTypeApplicationXCdf               string = "application/x-cdf"
	MediaTypeApplicationXCsh               string = "application/x-csh"
	MediaTypeApplicationXFontOtf           string = "application/x-font-otf"
	MediaTypeApplicationXFontTtf           string = "application/x-font-ttf"
	MediaTypeApplicationXFontWoff          string = "application/x-font-woff"
	MediaTypeApplicationXFreeArc           string = "application/x-freearc"
	MediaTypeApplicationXHttpdPhp          string = "application/x-httpd-php"
	MediaTypeApplicationXPkcs12            string = "application/x-pkcs12"
	MediaTypeApplicationXSh                string = "application/x-sh"
	MediaTypeApplicationXShockwaveFlash    string = "application/x-shockwave-flash"
	MediaTypeApplicationXSilverlightApp    string = "application/x-silverlight-app"
	MediaTypeApplicationXTar               string = "application/x-tar"
	MediaTypeApplicationXWwwFormUrlencoded string = "application/x-www-form-urlencoded"
)

// audit
const (
	MediaTypeAudioMidi     string = "audio/midi"
	MediaTypeAudioMP4      string = "audio/mp4"
	MediaTypeAudioMpeg     string = "audio/mpeg"
	MediaTypeAudioOgg      string = "audio/ogg"
	MediaTypeAudioOpus     string = "audio/opus"
	MediaTypeAudio3gpp     string = "audio/3gpp"
	MediaTypeAudio3gpp2    string = "audio/3gpp2"
	MediaTypeAudioWav      string = "audio/wav"
	MediaTypeAudioWebm     string = "audio/webm"
	MediaTypeAudioXAac     string = "audio/x-aac"
	MediaTypeAudioXAiff    string = "audio/x-aiff"
	MediaTypeAudioXMidi    string = "audio/x-midi"
	MediaTypeAudioXMpegURL string = "audio/x-mpegurl"
	MediaTypeAudioXMsWma   string = "audio/x-ms-wma"
	MediaTypeAudioXWav     string = "audio/x-wav"
)

// font
const (
	MediaTypeFontCollection string = "font/collection"
	MediaTypeFontOtf        string = "font/otf"
	MediaTypeFontSFnt       string = "font/sfnt"
	MediaTypeFontTtf        string = "font/ttf"
	MediaTypeFontWoff       string = "font/woff"
	MediaTypeFontWoff2      string = "font/woff2"
)

// image
const (
	MediaTypeImageAvif             string = "image/avif"
	MediaTypeImageBmp              string = "image/bmp"
	MediaTypeImageGif              string = "image/gif"
	MediaTypeImageJp2              string = "image/jp2"
	MediaTypeImageJpeg             string = "image/jpeg"
	MediaTypeImageJpm              string = "image/jpm"
	MediaTypeImageJpx              string = "image/jpx"
	MediaTypeImagePng              string = "image/png"
	MediaTypeImageSVGXML           string = "image/svg+xml"
	MediaTypeImageTiff             string = "image/tiff"
	MediaTypeImageVndMicrosoftIcon string = "image/vnd.microsoft.icon"
	MediaTypeImageWebp             string = "image/webp"
)

// multipart
const (
	MediaTypeMultipartByteRanges string = "multipart/byteranges"
	MediaTypeMultipartEncrypted  string = "multipart/encrypted"
	MediaTypeMultipartFormData   string = "multipart/form-data"
	MediaTypeMultipartRelated    string = "multipart/related"
)

// text
const (
	MediaTypeTextCalendar   string = "text/calendar"
	MediaTypeTextCSS        string = "text/css"
	MediaTypeTextCSV        string = "text/csv"
	MediaTypeTextHTML       string = "text/html"
	MediaTypeTextJavaScript string = "text/javascript"
	MediaTypeTextMarkdown   string = "text/markdown"
	MediaTypeTextPlain      string = "text/plain"
	MediaTypeTextRichText   string = "text/richtext"
	MediaTypeTextSGML       string = "text/sgml"
	MediaTypeTextXML        string = "text/xml"
	MediaTypeTextYAML       string = "text/yaml"
)

// video
const (
	MediaTypeVideoH264      string = "video/h264"
	MediaTypeVideoMJ2       string = "video/mj2"
	MediaTypeVideoMP2T      string = "video/mp2t"
	MediaTypeVideoMP4       string = "video/mp4"
	MediaTypeVideoMpeg      string = "video/mpeg"
	MediaTypeVideoOgg       string = "video/ogg"
	MediaTypeVideoQuicktime string = "video/quicktime"
	MediaTypeVideoThreegpp  string = "video/3gpp"
	MediaTypeVideoThreegpp2 string = "video/3gpp2"
	MediaTypeVideoWebm      string = "video/webm"
	MediaTypeVideoXMSVideo  string = "video/x-msvideo"
)
