import React, { useState, useEffect } from 'react';
import { Titlebar } from './components/Titlebar';
import { BackgroundImage } from './components/BackgroundImage';
import { ProfileSection } from './components/ProfileSection';
import { ControlSection } from './components/ControlSection';
import { MusicPlayer } from './components/MusicPlayer';
import { UpdateOverlay } from './components/UpdateOverlay';
import { ErrorModal } from './components/ErrorModal';
import { DeleteConfirmationModal } from './components/DeleteConfirmationModal';
import { ModManager } from './components/ModManager';
import hytaleLogo from './assets/logo.png';

import { 
  DownloadAndLaunch, 
  OpenFolder, 
  GetNick, 
  SetNick, 
  DeleteGame, 
  Update,
  ExitGame,
  IsGameRunning,
  // Version Manager
  GetVersionType,
  SetVersionType,
  GetSelectedVersion,
  SetSelectedVersion,
  GetVersionList,
  IsVersionInstalled,
  GetInstalledVersionsForBranch,
  CheckLatestNeedsUpdate,
  // Settings
  SelectInstanceDirectory
} from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const App: React.FC = () => {
  // User state
  const [username, setUsername] = useState<string>("HyPrism");
  const [isEditing, setIsEditing] = useState<boolean>(false);
  
  // Download state
  const [progress, setProgress] = useState<number>(0);
  const [isDownloading, setIsDownloading] = useState<boolean>(false);
  const [isGameRunning, setIsGameRunning] = useState<boolean>(false);
  const [downloaded, setDownloaded] = useState<number>(0);
  const [total, setTotal] = useState<number>(0);
  
  // Update state
  const [updateAsset, setUpdateAsset] = useState<any>(null);
  const [isUpdatingLauncher, setIsUpdatingLauncher] = useState<boolean>(false);
  const [updateStats, setUpdateStats] = useState({ d: 0, t: 0 });

  // Modal state
  const [showDelete, setShowDelete] = useState<boolean>(false);
  const [showModManager, setShowModManager] = useState<boolean>(false);
  const [error, setError] = useState<any>(null);

  // Version state
  const [currentBranch, setCurrentBranch] = useState<string>("release");
  const [currentVersion, setCurrentVersion] = useState<number>(0);
  const [availableVersions, setAvailableVersions] = useState<number[]>([]);
  const [installedVersions, setInstalledVersions] = useState<number[]>([]);
  const [isLoadingVersions, setIsLoadingVersions] = useState<boolean>(false);
  const [isVersionInstalled, setIsVersionInstalled] = useState<boolean>(false);
  const [isCheckingInstalled, setIsCheckingInstalled] = useState<boolean>(false);
  // const [customInstanceDir, setCustomInstanceDir] = useState<string>("");
  const [latestNeedsUpdate, setLatestNeedsUpdate] = useState<boolean>(false);

  // Check if current version is installed when branch or version changes
  useEffect(() => {
    const checkInstalled = async () => {
      if (currentVersion === 0) {
        // Version 0 is the auto-updating "latest" instance
        // Check if it's actually installed
        setIsCheckingInstalled(true);
        try {
          const installed = await IsVersionInstalled(currentBranch, 0);
          setIsVersionInstalled(installed);
          // Check if latest needs update
          const needsUpdate = await CheckLatestNeedsUpdate(currentBranch);
          setLatestNeedsUpdate(needsUpdate);
        } catch (e) {
          console.error('Failed to check latest instance:', e);
          setIsVersionInstalled(false);
          setLatestNeedsUpdate(false);
        }
        setIsCheckingInstalled(false);
        return;
      }
      if (currentVersion < 0) {
        // Uninitialized or invalid version
        setIsVersionInstalled(false);
        setLatestNeedsUpdate(false);
        return;
      }
      setIsCheckingInstalled(true);
      try {
        const installed = await IsVersionInstalled(currentBranch, currentVersion);
        setIsVersionInstalled(installed);
        setLatestNeedsUpdate(false); // Versioned instances don't auto-update
      } catch (e) {
        console.error('Failed to check if version installed:', e);
        setIsVersionInstalled(false);
        setLatestNeedsUpdate(false);
      }
      setIsCheckingInstalled(false);
    };
    checkInstalled();
  }, [currentBranch, currentVersion]);

  // Load version list when branch changes
  useEffect(() => {
    const loadVersions = async () => {
      setIsLoadingVersions(true);
      try {
        const versions = await GetVersionList(currentBranch);
        setAvailableVersions(versions || []);
        
        // Load installed versions
        const installed = await GetInstalledVersionsForBranch(currentBranch);
        setInstalledVersions(installed || []);
        
        // If current version is not valid for this branch, set to latest
        if (versions && !versions.includes(currentVersion) && versions.length > 0) {
          setCurrentVersion(versions[0]);
          await SetSelectedVersion(versions[0]);
        }
      } catch (e) {
        console.error('Failed to load versions:', e);
        setAvailableVersions([]);
        setInstalledVersions([]);
      }
      setIsLoadingVersions(false);
    };
    loadVersions();
  }, [currentBranch]);

  // Handle branch change - immediately load and set latest version for new branch
  const handleBranchChange = async (branch: string) => {
    setCurrentBranch(branch);
    await SetVersionType(branch);
    
    // Load versions for new branch and set to latest (version 0)
    setIsLoadingVersions(true);
    try {
      const versions = await GetVersionList(branch);
      setAvailableVersions(versions);
      // Always set to "latest" (version 0) when switching branches
      setCurrentVersion(0);
      await SetSelectedVersion(0);
    } catch (e) {
      console.error('Failed to load versions for branch:', e);
    }
    setIsLoadingVersions(false);
  };

  // Handle version change
  const handleVersionChange = async (version: number) => {
    setCurrentVersion(version);
    await SetSelectedVersion(version);
  };

  // Game state polling
  useEffect(() => {
    if (!isGameRunning) return;
    
    const pollInterval = setInterval(async () => {
      try {
        const running = await IsGameRunning();
        if (!running) {
          setIsGameRunning(false);
          setProgress(0);
        }
      } catch (e) {
        console.error('Failed to check game state:', e);
      }
    }, 3000); // Check every 3 seconds (reduced frequency)
    
    return () => clearInterval(pollInterval);
  }, [isGameRunning]);

  useEffect(() => {
    // Initialize user settings
    GetNick().then((n: string) => n && setUsername(n));
    // GetCustomInstanceDir().then((dir: string) => setCustomInstanceDir(dir));
    
    // Load saved branch and version - must load branch first, then version
    const loadSettings = async () => {
      try {
        // Get saved branch (defaults to "release" in backend if not set)
        const savedBranch = await GetVersionType();
        const branch = savedBranch || "release";
        setCurrentBranch(branch);
        
        // Load version list for this branch
        setIsLoadingVersions(true);
        const versions = await GetVersionList(branch);
        setAvailableVersions(versions);
        
        // Load installed versions
        const installed = await GetInstalledVersionsForBranch(branch);
        setInstalledVersions(installed);
        
        // Always start with "latest" (version 0) when launcher opens
        setCurrentVersion(0);
        await SetSelectedVersion(0);
        setIsLoadingVersions(false);
      } catch (e) {
        console.error('Failed to load settings:', e);
        setIsLoadingVersions(false);
      }
    };
    loadSettings();

    // Event listeners
    const unsubProgress = EventsOn('progress-update', (data: any) => {
      setProgress(data.progress);
      setDownloaded(data.downloaded);
      setTotal(data.total);
      
      if (data.progress >= 100 && data.stage === 'launch') {
        setIsGameRunning(true);
        setIsDownloading(false);
      }
    });

    const unsubUpdate = EventsOn('update:available', (asset: any) => {
      setUpdateAsset(asset);
      // Don't auto-update - let user click the update button
      console.log('Update available:', asset);
    });

    const unsubUpdateProgress = EventsOn('update:progress', (_stage: string, progress: number, _message: string, _file: string, _speed: string, downloaded: number, total: number) => {
      setProgress(progress);
      setUpdateStats({ d: downloaded, t: total });
    });

    const unsubError = EventsOn('error', (err: any) => {
      setError(err);
      setIsDownloading(false);
    });

    return () => {
      unsubProgress();
      unsubUpdate();
      unsubUpdateProgress();
      unsubError();
    };
  }, []);

  const handleUpdate = async () => {
    setIsUpdatingLauncher(true);
    setProgress(0);
    setUpdateStats({ d: 0, t: 0 });
    
    try {
      await Update();
    } catch (err) {
      console.error('Update failed:', err);
      setError({
        type: 'UPDATE_ERROR',
        message: 'Failed to update launcher',
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString()
      });
      setIsUpdatingLauncher(false);
    }
  };

  const handlePlay = async () => {
    if (!username.trim() || username.length > 16) {
      setError({
        type: 'VALIDATION',
        message: 'Invalid Nickname',
        technical: 'Nickname must be between 1 and 16 characters',
        timestamp: new Date().toISOString()
      });
      return;
    }
    
    setIsDownloading(true);
    try {
      await DownloadAndLaunch(username);
    } catch (err) {
      console.error('Launch failed:', err);
      setIsDownloading(false);
    }
  };

  const handleNickChange = async (newNick: string) => {
    setUsername(newNick);
    await SetNick(newNick);
  };

  const handleExit = async () => {
    try {
      await ExitGame();
    } catch (err) {
      console.error('Failed to exit game:', err);
    }
    setIsGameRunning(false);
    setProgress(0);
  };

  const handleCustomDirChange = async () => {
    try {
      const selectedDir = await SelectInstanceDirectory();
      if (selectedDir) {
        console.log('Instance directory updated to:', selectedDir);
        
        // Reload version list and check installed versions for new directory
        setIsLoadingVersions(true);
        try {
          const versions = await GetVersionList(currentBranch);
          setAvailableVersions(versions || []);
          
          const installed = await GetInstalledVersionsForBranch(currentBranch);
          setInstalledVersions(installed || []);
          
          // Re-check if current version is installed in new directory
          const isInstalled = await IsVersionInstalled(currentBranch, currentVersion);
          setIsVersionInstalled(isInstalled);
          
          // Check if latest needs update
          if (currentVersion === 0) {
            const needsUpdate = await CheckLatestNeedsUpdate(currentBranch);
            setLatestNeedsUpdate(needsUpdate);
          }
        } catch (e) {
          console.error('Failed to reload versions after directory change:', e);
        }
        setIsLoadingVersions(false);
      }
    } catch (err) {
      console.error('Failed to change instance directory:', err);
      setError({
        type: 'SETTINGS_ERROR',
        message: 'Failed to change instance directory',
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString()
      });
    }
  };

  return (
    <div className="relative w-screen h-screen bg-[#090909] text-white overflow-hidden font-sans select-none">
      <BackgroundImage />
      <Titlebar />
      
      {/* Music Player - positioned in top right */}
      <div className="absolute top-12 right-4 z-20">
        <MusicPlayer forceMuted={isGameRunning} />
      </div>
      
      {isUpdatingLauncher && (
        <UpdateOverlay 
          progress={progress} 
          downloaded={updateStats.d} 
          total={updateStats.t} 
        />
      )}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection 
            username={username}
            isEditing={isEditing}
            onEditToggle={setIsEditing}
            onUserChange={handleNickChange}
            updateAvailable={!!updateAsset}
            onUpdate={handleUpdate}
          />

          {/* Hytale Logo - Right Side */}
          <img src={hytaleLogo} alt="Hytale" className="h-24 drop-shadow-2xl" />
        </div>

        <ControlSection 
          onPlay={handlePlay}
          onDownload={handlePlay}
          onExit={handleExit}
          isDownloading={isDownloading}
          isGameRunning={isGameRunning}
          isVersionInstalled={isVersionInstalled}
          isCheckingInstalled={isCheckingInstalled}
          progress={progress}
          downloaded={downloaded}
          total={total}
          currentBranch={currentBranch}
          currentVersion={currentVersion}
          availableVersions={availableVersions}
          installedVersions={installedVersions}
          isLoadingVersions={isLoadingVersions}
          latestNeedsUpdate={latestNeedsUpdate}
          onBranchChange={handleBranchChange}
          onVersionChange={handleVersionChange}
          onCustomDirChange={handleCustomDirChange}
          actions={{
            openFolder: OpenFolder,
            showDelete: () => setShowDelete(true),
            showModManager: () => setShowModManager(true)
          }}
        />
      </main>

      {/* Modals */}
      {showDelete && (
        <DeleteConfirmationModal 
          onConfirm={() => { 
            DeleteGame(); 
            setShowDelete(false); 
          }} 
          onCancel={() => setShowDelete(false)} 
        />
      )}
      
      {error && (
        <ErrorModal 
          error={error} 
          onClose={() => setError(null)} 
        />
      )}

      {showModManager && (
        <ModManager
          onClose={() => setShowModManager(false)}
          currentBranch={currentBranch}
          currentVersion={currentVersion}
        />
      )}
    </div>
  );
};

export default App;
