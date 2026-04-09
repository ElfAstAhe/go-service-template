package transport

// application
const (
	MimeTypeApplicationAtomXML      string = "application/atom+xml"
	MimeTypeApplicationAtomCatXML   string = "application/atomcat+xml"
	MimeTypeApplicationEcmaScript   string = "application/ecmascript"
	MimeTypeApplicationEpubZip      string = "application/epub+zip"
	MimeTypeApplicationGZip         string = "application/gzip"
	MimeTypeApplicationJavaArchive  string = "application/java-archive"
	MimeTypeApplicationJavaScript   string = "application/javascript"
	MimeTypeApplicationJSON         string = "application/json"
	MimeTypeApplicationLdJSON       string = "application/ld+json"
	MimeTypeApplicationManifestJSON string = "application/manifest+json"
	MimeTypeApplicationMP4          string = "application/mp4"
	MimeTypeApplicationMSWord       string = "application/msword"
	MimeTypeApplicationOctetStream  string = "application/octet-stream"
	MimeTypeApplicationOgg          string = "application/ogg"
	MimeTypeApplicationPDF          string = "application/pdf"
	MimeTypeApplicationPkcs10       string = "application/pkcs10"
	MimeTypeApplicationPkcs7Mime    string = "application/pkcs7-mime"
	MimeTypeApplicationPkcs7Sig     string = "application/pkcs7-signature"
	MimeTypeApplicationPkcs8        string = "application/pkcs8"
	MimeTypeApplicationPostScript   string = "application/postscript"
	MimeTypeApplicationRdfXML       string = "application/rdf+xml"
	MimeTypeApplicationRssXML       string = "application/rss+xml"
	MimeTypeApplicationRTF          string = "application/rtf"
	MimeTypeApplicationSmilXML      string = "application/smil+xml"
	MimeTypeApplicationXhtmlXML     string = "application/xhtml+xml"
	MimeTypeApplicationXML          string = "application/xml"
	MimeTypeApplicationXmlDTD       string = "application/xml-dtd"
	MimeTypeApplicationXsltXML      string = "application/xslt+xml"
	MimeTypeApplicationZip          string = "application/zip"

	// Vendor specific
	MimeTypeApplicationVndAmazonEbook          string = "application/vnd.amazon.ebook"
	MimeTypeApplicationVndAppleInstallerXML    string = "application/vnd.apple.installer+xml"
	MimeTypeApplicationVndMozillaXulXML        string = "application/vnd.mozilla.xul+xml"
	MimeTypeApplicationVndMSExcel              string = "application/vnd.ms-excel"
	MimeTypeApplicationVndMSFontObject         string = "application/vnd.ms-fontobject"
	MimeTypeApplicationVndMSPowerpoint         string = "application/vnd.ms-powerpoint"
	MimeTypeApplicationVndOasisDocPresentation string = "application/vnd.oasis.opendocument.presentation"
	MimeTypeApplicationVndOasisDocSpreadsheet  string = "application/vnd.oasis.opendocument.spreadsheet"
	MimeTypeApplicationVndOasisDocText         string = "application/vnd.oasis.opendocument.text"

	// OpenXML
	MimeTypeApplicationVndOpenXMLDocPresentation string = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MimeTypeApplicationVndOpenXMLDocSpreadsheet  string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MimeTypeApplicationVndOpenXMLDocWord         string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	// X-types
	MimeTypeApplicationVndRar             string = "application/vnd.rar"
	MimeTypeApplicationVndVisio           string = "application/vnd.visio"
	MimeTypeApplicationX7zCompressed      string = "application/x-7z-compressed"
	MimeTypeApplicationXAbiword           string = "application/x-abiword"
	MimeTypeApplicationXBZip              string = "application/x-bzip"
	MimeTypeApplicationXBZip2             string = "application/x-bzip2"
	MimeTypeApplicationXCdf               string = "application/x-cdf"
	MimeTypeApplicationXCsh               string = "application/x-csh"
	MimeTypeApplicationXFontOtf           string = "application/x-font-otf"
	MimeTypeApplicationXFontTtf           string = "application/x-font-ttf"
	MimeTypeApplicationXFontWoff          string = "application/x-font-woff"
	MimeTypeApplicationXFreeArc           string = "application/x-freearc"
	MimeTypeApplicationXHttpdPhp          string = "application/x-httpd-php"
	MimeTypeApplicationXPkcs12            string = "application/x-pkcs12"
	MimeTypeApplicationXSh                string = "application/x-sh"
	MimeTypeApplicationXShockwaveFlash    string = "application/x-shockwave-flash"
	MimeTypeApplicationXSilverlightApp    string = "application/x-silverlight-app"
	MimeTypeApplicationXTar               string = "application/x-tar"
	MimeTypeApplicationXWwwFormUrlencoded string = "application/x-www-form-urlencoded"
)

// audit
const (
	MimeTypeAudioMidi     string = "audio/midi"
	MimeTypeAudioMP4      string = "audio/mp4"
	MimeTypeAudioMpeg     string = "audio/mpeg"
	MimeTypeAudioOgg      string = "audio/ogg"
	MimeTypeAudioOpus     string = "audio/opus"
	MimeTypeAudio3gpp     string = "audio/3gpp"
	MimeTypeAudio3gpp2    string = "audio/3gpp2"
	MimeTypeAudioWav      string = "audio/wav"
	MimeTypeAudioWebm     string = "audio/webm"
	MimeTypeAudioXAac     string = "audio/x-aac"
	MimeTypeAudioXAiff    string = "audio/x-aiff"
	MimeTypeAudioXMidi    string = "audio/x-midi"
	MimeTypeAudioXMpegURL string = "audio/x-mpegurl"
	MimeTypeAudioXMsWma   string = "audio/x-ms-wma"
	MimeTypeAudioXWav     string = "audio/x-wav"
)

// font
const (
	MimeTypeFontCollection string = "font/collection"
	MimeTypeFontOtf        string = "font/otf"
	MimeTypeFontSFnt       string = "font/sfnt"
	MimeTypeFontTtf        string = "font/ttf"
	MimeTypeFontWoff       string = "font/woff"
	MimeTypeFontWoff2      string = "font/woff2"
)

// image
const (
	MimeTypeImageAvif             string = "image/avif"
	MimeTypeImageBmp              string = "image/bmp"
	MimeTypeImageGif              string = "image/gif"
	MimeTypeImageJp2              string = "image/jp2"
	MimeTypeImageJpeg             string = "image/jpeg"
	MimeTypeImageJpm              string = "image/jpm"
	MimeTypeImageJpx              string = "image/jpx"
	MimeTypeImagePng              string = "image/png"
	MimeTypeImageSVGXML           string = "image/svg+xml"
	MimeTypeImageTiff             string = "image/tiff"
	MimeTypeImageVndMicrosoftIcon string = "image/vnd.microsoft.icon"
	MimeTypeImageWebp             string = "image/webp"
)

// multipart
const (
	MimeTypeMultipartByteRanges string = "multipart/byteranges"
	MimeTypeMultipartEncrypted  string = "multipart/encrypted"
	MimeTypeMultipartFormData   string = "multipart/form-data"
	MimeTypeMultipartRelated    string = "multipart/related"
)

// text
const (
	MimeTypeTextCalendar   string = "text/calendar"
	MimeTypeTextCSS        string = "text/css"
	MimeTypeTextCSV        string = "text/csv"
	MimeTypeTextHTML       string = "text/html"
	MimeTypeTextJavaScript string = "text/javascript"
	MimeTypeTextMarkdown   string = "text/markdown"
	MimeTypeTextPlain      string = "text/plain"
	MimeTypeTextRichText   string = "text/richtext"
	MimeTypeTextSGML       string = "text/sgml"
	MimeTypeTextXML        string = "text/xml"
	MimeTypeTextYAML       string = "text/yaml"
)

// video
const (
	MimeTypeVideoH264      string = "video/h264"
	MimeTypeVideoMJ2       string = "video/mj2"
	MimeTypeVideoMP2T      string = "video/mp2t"
	MimeTypeVideoMP4       string = "video/mp4"
	MimeTypeVideoMpeg      string = "video/mpeg"
	MimeTypeVideoOgg       string = "video/ogg"
	MimeTypeVideoQuicktime string = "video/quicktime"
	MimeTypeVideoThreegpp  string = "video/3gpp"
	MimeTypeVideoThreegpp2 string = "video/3gpp2"
	MimeTypeVideoWebm      string = "video/webm"
	MimeTypeVideoXMSVideo  string = "video/x-msvideo"
)
