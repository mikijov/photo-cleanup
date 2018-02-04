photo-cleanup
=============

[![License][License-Image]][License-Url]
[![Build][Build-Status-Image]][Build-Status-Url]
[![ReportCard][ReportCard-Image]][ReportCard-Url]

photo-cleanup is a photo organizer.

    $ photo-cleanup help organize
    Moves photos from source into proper destination subdirectory.

    Usage:
      photo-cleanup organize srcdir destdir [flags]

    Flags:
          --all-files                   Process all files. Default is only images (jpg).
          --dir-fmt string              Directory format (default "yyyy/mm")
      -n, --dry-run                     Do not make any changes to files, only show what would happen.
      -h, --help                        help for organize
          --hidden-files                Process hidden files. Default is only normal files.
          --min-size int                Minimum file size to consider for processing.
          --use-exif-time               Use time from exif meta data. (default true)
          --use-file-time               Use file modification time when no meta data.
          --use-filename-encoded-time   Attempt to parse time from filename. (default true)

    Global Flags:
          --config string   config file (default is $HOME/.photo-cleanup.yaml)
      -q, --quiet           display no information while processing
      -v, --verbose         display more information while processing

To move all images from source to target directory while at the same time
organizing them into /dest/<YEAR>/<MONTH>/filename.ext structure, simply execute:

    $ photo-cleanup organize /media/SDCARD /home/me/Photos

photo-cleanup is configured with common defaults, to ignore all hidden and
non-image files. However, it supports multiple options to change destination
directory structure, allow hidden and non-image files and many others.

General algorithm is as follows:
- find all jpg, jpeg and mp4 files
- determine creation time:
  - if file has embedded exif metadata use it (exif2 library currently supports
    only jpeg files)
  - if no exif data, see if the filename is in the IMG_yyyymmdd_HHMMSS.jpg or
  VID_yyyymmdd_HHMMSS.mp4 format and if so, extract the date.
  - if still no date and if --use-file-time is set, use file modification time
- create new filepath using yyyy/mm or format specified using --dir-fmt
- move all prepared files into new destination, skipping any files that already
  exist

## Features and ToDo
- [x] extract date/time from jpegs files
- [x] allow to customize destination directory format
- [ ] organize duplicate filenames by appending -1, -2 etc.
- [ ] detect binary identical files
- [ ] extract date/time from mp4 files
- [ ] support other file formats
- [ ] organize using hard links instead of moving files

## No Warranty

Please note that photo-cleanup comes with no warranty. I use it to manage my
photos, but I am sure that some bugs have sneaked through. Backup your photos
before letting photo-cleanup lose. See license for details.

## Thanks

Thanks to Bobi Jones whose
[jpeg](http://www.publicdomainpictures.net/view-image.php?image=22282) is used
as test data.

[License-Url]: https://opensource.org/licenses/Apache-2.0
[License-Image]: https://img.shields.io/badge/license-Apache%202.0-blue.svg?maxAge=2592000
[Build-Status-Url]: http://travis-ci.org/mikijov/photo-cleanup
[Build-Status-Image]: https://travis-ci.org/mikijov/photo-cleanup.svg?branch=master
[ReportCard-Url]: https://goreportcard.com/report/github.com/mikijov/photo-cleanup
[ReportCard-Image]: https://goreportcard.com/badge/github.com/mikijov/photo-cleanup
