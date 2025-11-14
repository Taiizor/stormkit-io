package integrations

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"go.uber.org/zap"
)

const removeAfterInactivity = time.Hour * 6

type DownloadFunc func(deploymentID, bucketname, keyprefix string) (string, error)

type Zip struct {
	Location string
	cacheKey string
	timer    *time.Timer
	manager  *ZipManager
}

// NewZip creates a new Zip instance and attaches a timer that will remove
// the folder after the configured time.
func NewZip(key string, manager *ZipManager) *Zip {
	zip := &Zip{cacheKey: key, manager: manager}
	zip.timer = time.AfterFunc(removeAfterInactivity, zip.RemoveFolder)

	return zip
}

// RemoveFolder removes the folder from the file system and also deletes
// itself from the cache.
func (z *Zip) RemoveFolder() {
	z.manager.mux.Lock()
	defer z.manager.mux.Unlock()

	slog.Debug(slog.LogOpts{
		Msg:   "Removing folder after inactivity",
		Level: slog.DL2,
		Payload: []zap.Field{
			zap.String("folder", z.Location),
			zap.String("cacheKey", z.cacheKey),
		},
	})

	if z.Location != "" {
		if err := os.RemoveAll(z.Location); err != nil {
			slog.Errorf("error while removing zip folder: %s", err.Error())
		}
	}

	delete(z.manager.cache, z.cacheKey)
}

type ZipManager struct {
	download DownloadFunc
	cache    map[string]*Zip // deployment id => location
	mux      sync.Mutex
}

func NewZipManager(download DownloadFunc) *ZipManager {
	return &ZipManager{
		cache:    map[string]*Zip{},
		download: download,
	}
}

func (zm *ZipManager) folderDoesNotExist(folder string) bool {
	if folder == "" {
		return true
	}

	_, err := os.Stat(folder)
	return os.IsNotExist(err)
}

// GetFile return a file for the given deployment and location. The location is
// the fully qualified location name (e.g. aws:/bucket-name/key-prefix)
// This method will download the zip if necessary. The downloader function is configured
// when calling the `NewZipManager` method.
func (zm *ZipManager) GetFile(args GetFileArgs) (*GetFileResult, error) {
	zm.mux.Lock()
	defer zm.mux.Unlock()

	var err error

	did := args.DeploymentID.String()
	pieces := strings.Split(strings.TrimPrefix(args.Location, "aws:"), "/")
	bucket := pieces[0]
	prefix := filepath.Join(pieces[1:]...) // Relative path

	if zm.cache[did] == nil || zm.folderDoesNotExist(zm.cache[did].Location) {
		zm.cache[did] = NewZip(did, zm)
		zm.cache[did].Location, err = zm.download(did, bucket, prefix)

		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to download zip file").
				WithContext("deployment_id", did).
				WithContext("bucket", bucket).
				WithContext("prefix", prefix)
		}
	}

	zm.cache[did].timer.Reset(removeAfterInactivity)

	filePath := path.Join(zm.cache[did].Location, args.FileName)
	data, err := os.ReadFile(filePath)

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to read file from zip").
			WithContext("file_path", filePath).
			WithContext("deployment_id", did).
			WithContext("file_name", args.FileName)
	}

	stat, err := os.Stat(filePath)

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to stat file from zip").
			WithContext("file_path", filePath).
			WithContext("deployment_id", did).
			WithContext("file_name", args.FileName)
	}

	return &GetFileResult{
		Content:     data,
		ContentType: DetectContentType(filePath, data),
		Size:        stat.Size(),
	}, nil
}
