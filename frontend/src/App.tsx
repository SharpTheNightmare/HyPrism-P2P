import React, { useState, useEffect } from 'react';
import { Titlebar } from './components/Titlebar';
import { BackgroundImage } from './components/BackgroundImage';
import { ProfileSection } from './components/ProfileSection';
import { ControlSection } from './components/ControlSection';
import { MusicPlayer } from './components/MusicPlayer';
import { UpdateOverlay } from './components/UpdateOverlay';
import { ErrorModal } from './components/ErrorModal';
import { ModManager } from './components/ModManager';
import { AuthModal } from './components/AuthModal';
import hytaleLogo from './assets/logo.png';

import {
  Launch,
  Update,
  ExitGame,
  IsGameRunning,
  // Settings
  SelectGameInstallDirectory,
  GetNews,
  GetLauncherVersion,
  // Authentication
  LoginWithHytaleAccount,
  LogoutHytaleAccount,
  GetAuthStatus,
  OpenAuthURL,
} from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { NewsPreview } from './components/NewsPreview';

const App: React.FC = () => {
  // User state
  const [username, setUsername] = useState<string>("HyPrism");
  const [launcherVersion, setLauncherVersion] = useState<string>("dev");

  // Game state
  const [isGameRunning, setIsGameRunning] = useState<boolean>(false);

  // Update state
  const [updateAsset, setUpdateAsset] = useState<any>(null);
  const [isUpdatingLauncher, setIsUpdatingLauncher] = useState<boolean>(false);
  const [updateProgress, setUpdateProgress] = useState<number>(0);
  const [updateStats, setUpdateStats] = useState({ d: 0, t: 0 });

  // Modal state
  const [showModManager, setShowModManager] = useState<boolean>(false);
  const [error, setError] = useState<any>(null);

  // Auth state
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(false);
  const [userUUID, setUserUUID] = useState<string>('');
  const [showAuthModal, setShowAuthModal] = useState<boolean>(false);
  const [authMessage, setAuthMessage] = useState<string>('');
  const [isAuthenticating, setIsAuthenticating] = useState<boolean>(false);



  // Game state polling
  useEffect(() => {
    if (!isGameRunning) return;

    const pollInterval = setInterval(async () => {
      try {
        const running = await IsGameRunning();
        if (!running) {
          setIsGameRunning(false);
        }
      } catch (e) {
        console.error('Failed to check game state:', e);
      }
    }, 3000);

    return () => clearInterval(pollInterval);
  }, [isGameRunning]);

  useEffect(() => {
    // Check auth status on startup
    const checkAuth = async () => {
      try {
        const authStatus = await GetAuthStatus();
        if (authStatus.logged_in) {
          setIsLoggedIn(true);
          setUsername(authStatus.username || 'HyPrism');
          setUserUUID(authStatus.uuid || '');
        }
      } catch (err) {
        console.error('Failed to check auth status:', err);
      }
    };
    checkAuth();

    // Initialize launcher version
    GetLauncherVersion().then((v: string) => setLauncherVersion(v));

    // Event listeners
    const unsubProgress = EventsOn('progress-update', (data: any) => {
      if (data.stage === 'launch') {
        setIsGameRunning(true);
      }
    });

    const unsubUpdate = EventsOn('update:available', (asset: any) => {
      setUpdateAsset(asset);
      // Don't auto-update - let user click the update button
      console.log('Update available:', asset);
    });

    const unsubUpdateProgress = EventsOn('update:progress', (_stage: string, progress: number, _message: string, _file: string, _speed: string, downloaded: number, total: number) => {
      setUpdateProgress(progress);
      setUpdateStats({ d: downloaded, t: total });
    });

    const unsubError = EventsOn('error', (err: any) => {
      setError(err);
    });

    // Auth event listeners
    const unsubAuthProgress = EventsOn('auth-progress', (message: string) => {
      console.log('Auth progress:', message);
      setAuthMessage(message);
    });

    const unsubAuthSuccess = EventsOn('auth-success', (data: any) => {
      console.log('Auth success:', data);
      setIsLoggedIn(true);
      setUsername(data.username);
      setUserUUID(data.uuid);
      setAuthMessage(`Logged in as: ${data.username}`);
      setIsAuthenticating(false);
      // Close modal after a delay to show success message
      setTimeout(() => {
        setShowAuthModal(false);
        setAuthMessage('');
      }, 2000);
    });

    const unsubAuthLogout = EventsOn('auth-logout', () => {
      console.log('Logged out');
      setIsLoggedIn(false);
      setUserUUID('');
      setUsername('HyPrism');
    });

    return () => {
      unsubProgress();
      unsubUpdate();
      unsubUpdateProgress();
      unsubError();
      unsubAuthProgress();
      unsubAuthSuccess();
      unsubAuthLogout();
    };
  }, []);

  const handleUpdate = async () => {
    setIsUpdatingLauncher(true);
    setUpdateProgress(0);
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

    try {
      await Launch(username);
    } catch (err) {
      console.error('Launch failed:', err);
    }
  };

  const handleExit = async () => {
    try {
      await ExitGame();
    } catch (err) {
      console.error('Failed to exit game:', err);
    }
    setIsGameRunning(false);
  };

  const handleLogin = async () => {
    try {
      setShowAuthModal(true);
      setIsAuthenticating(true);
      setAuthMessage('');
      await LoginWithHytaleAccount();
    } catch (err) {
      console.error('Login failed:', err);
      setIsAuthenticating(false);
      setShowAuthModal(false);
      setError({
        type: 'AUTH_ERROR',
        message: 'Failed to login',
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString()
      });
    }
  };

  const handleLogout = async () => {
    try {
      await LogoutHytaleAccount();
      setIsLoggedIn(false);
      setUserUUID('');
      setUsername('HyPrism');
    } catch (err) {
      console.error('Logout failed:', err);
      setError({
        type: 'AUTH_ERROR',
        message: 'Failed to logout',
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString()
      });
    }
  };

  const handleOpenAuthURL = async (url: string) => {
    try {
      await OpenAuthURL(url);
    } catch (err) {
      console.error('Failed to open auth URL:', err);
    }
  };

  const handleGameLocationChange = async () => {
    try {
      const selectedDir = await SelectGameInstallDirectory();
      if (selectedDir) {
        console.log('Game install directory updated to:', selectedDir);
        setError({
          type: 'INFO',
          message: 'Game Location Updated',
          technical: `HyPrism will now use the official Hytale installation at:\n${selectedDir}`,
          timestamp: new Date().toISOString()
        });
      }
    } catch (err) {
      console.error('Failed to change game install directory:', err);
      setError({
        type: 'SETTINGS_ERROR',
        message: 'Failed to change game install directory',
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
          progress={updateProgress}
          downloaded={updateStats.d}
          total={updateStats.t}
        />
      )}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection
            username={username}
            updateAvailable={!!updateAsset}
            onUpdate={handleUpdate}
            launcherVersion={launcherVersion}
            isLoggedIn={isLoggedIn}
            onLogin={handleLogin}
            onLogout={handleLogout}
            uuid={userUUID}
          />
          {/* Hytale Logo & News - Right Side */}
          <div className="flex flex-col items-end gap-3">
            <img src={hytaleLogo} alt="Hytale" className="h-24 drop-shadow-2xl" />
            <NewsPreview getNews={async () => await GetNews(3)} />
          </div>
        </div>

        <ControlSection
          onPlay={handlePlay}
          onExit={handleExit}
          isGameRunning={isGameRunning}
          isLoggedIn={isLoggedIn}
          onGameLocationChange={handleGameLocationChange}
          actions={{
            showModManager: () => setShowModManager(true)
          }}
        />
      </main>

      {/* Modals */}
      {error && (
        <ErrorModal
          error={error}
          onClose={() => setError(null)}
        />
      )}

      {showModManager && (
        <ModManager
          onClose={() => setShowModManager(false)}
          currentBranch="release"
          currentVersion={0}
        />
      )}

      {/* Auth Modal */}
      <AuthModal
        isOpen={showAuthModal}
        onClose={() => {
          setShowAuthModal(false);
          setIsAuthenticating(false);
          setAuthMessage('');
        }}
        authMessage={authMessage}
        isAuthenticating={isAuthenticating}
        onOpenURL={handleOpenAuthURL}
      />
    </div>
  );
};

export default App;
