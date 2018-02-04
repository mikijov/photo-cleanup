photo-cleanup
=============

photo-cleanup is a photo organizer.

    $>photo-cleanup help organize
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

    photo-cleanup organize /media/SDCARD /home/me/Photos

photo-cleanup is configured with common defaults, to ignore all hidden and
non-image files. However, it supports multiple options to change destination
directory structure, allow hidden and non-image files and many others.

# No Warranty

Please note that photo-cleanup comes with no warranty. I use it to manage my
photos, but I am sure that some bugs have sneaked through. Backup your photos
before letting photo-cleanup lose. See license for details.

# Thanks

Thanks to Bobi Jones whose jpeg is used as test data.
http://www.publicdomainpictures.net/view-image.php?image=22282
