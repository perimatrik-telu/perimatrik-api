// migrate-media copies the recovered public disk into Cloudinary without changing names.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("migrate-media", flag.ContinueOnError)
	flags.SetOutput(out)
	source := flags.String("source", "./storage/app/public", "recovered Laravel public disk")
	apply := flags.Bool("apply", false, "upload to Cloudinary; default only lists files")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	info, err := os.Lstat(*source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source must be a directory, not a symlink")
	}
	var media *storage.Cloudinary
	if *apply {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		media, err = storage.NewCloudinary(cfg.Cloudinary)
		if err != nil {
			return err
		}
	}
	var processed, planned int
	for _, dir := range []string{storage.ArticleDir, storage.GalleryDir, storage.NewsDir, storage.AvatarDir, storage.MemberDir} {
		root := filepath.Join(*source, dir)
		if _, err := os.Lstat(root); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		err := filepath.WalkDir(root, func(p string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink refused: %s", p)
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("non-regular file refused: %s", p)
			}
			if entry.Name() == ".gitignore" || entry.Name() == ".DS_Store" {
				return nil
			}
			rel, err := filepath.Rel(*source, p)
			if err != nil {
				return err
			}
			key := filepath.ToSlash(rel)
			if _, ok := storage.MediaKey(key); !ok {
				return fmt.Errorf("unsupported media file: %s", key)
			}
			if !*apply {
				planned++
				fmt.Fprintf(out, "Would upload %s\n", key)
				return nil
			}
			opCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			file, err := os.Open(p)
			if err != nil {
				return err
			}
			defer file.Close()
			if err := media.Put(opCtx, key, file); err != nil {
				return fmt.Errorf("upload %s: %w", key, err)
			}
			processed++
			fmt.Fprintf(out, "Processed %s (existing assets are retained)\n", key)
			return nil
		})
		if err != nil {
			return err
		}
	}
	if *apply {
		fmt.Fprintf(out, "Processed %d files; existing assets retained. No database changes.\n", processed)
	} else {
		fmt.Fprintf(out, "Dry run: %d files. Use --apply to upload. No database changes.\n", planned)
	}
	return nil
}
