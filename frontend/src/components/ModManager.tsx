import React, { useState, useEffect, useCallback, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  X, Search, Download, Trash2, FolderOpen, 
  Package, ToggleLeft, ToggleRight, ExternalLink,
  Loader2, AlertCircle, RefreshCw, ArrowUp, TrendingUp, Clock
} from 'lucide-react';

interface Mod {
  id: string;
  name: string;
  version: string;
  author: string;
  description: string;
  enabled: boolean;
  iconUrl?: string;
  downloads?: number;
  category?: string;
  curseForgeId?: number;
}

interface CurseForgeMod {
  id: number;
  name: string;
  slug: string;
  summary: string;
  downloadCount: number;
  logo?: { thumbnailUrl: string; url: string };
  screenshots?: { id: number; title: string; thumbnailUrl: string; url: string }[];
  authors: { name: string }[];
  categories: { name: string }[];
  dateModified?: string;
  dateReleased?: string;
}

interface ModManagerProps {
  onClose: () => void;
  searchMods: (query: string, categoryId: number, page: number) => Promise<{ mods: CurseForgeMod[]; totalCount: number }>;
  getInstalledMods: () => Promise<Mod[]>;
  installMod: (modId: number) => Promise<void>;
  uninstallMod: (modId: string) => Promise<void>;
  toggleMod: (modId: string, enabled: boolean) => Promise<void>;
  getModCategories: () => Promise<{ id: number; name: string }[]>;
  openModsFolder: () => Promise<void>;
  checkModUpdates?: () => Promise<{ modId: string; hasUpdate: boolean; newVersion: string }[]>;
  getModDetails?: (modId: number) => Promise<CurseForgeMod>;
}

const formatDownloads = (count: number): string => {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
  return count.toString();
};

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
  return `${Math.floor(diffDays / 365)} years ago`;
};

type SortOption = 'name' | 'downloads' | 'updated';

export const ModManager: React.FC<ModManagerProps> = ({
  onClose,
  searchMods,
  getInstalledMods,
  installMod,
  uninstallMod,
  toggleMod,
  getModCategories,
  openModsFolder,
  checkModUpdates,
  getModDetails,
}) => {
  const [activeTab, setActiveTab] = useState<'browse' | 'installed'>('installed');
  const [searchQuery, setSearchQuery] = useState('');
  const [installedMods, setInstalledMods] = useState<Mod[]>([]);
  const [searchResults, setSearchResults] = useState<CurseForgeMod[]>([]);
  const [categories, setCategories] = useState<{ id: number; name: string }[]>([]);
  const [selectedCategory, setSelectedCategory] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isInstalling, setIsInstalling] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<SortOption>('downloads');
  const [modUpdates, setModUpdates] = useState<Record<string, { hasUpdate: boolean; newVersion: string }>>({});
  const [isCheckingUpdates, setIsCheckingUpdates] = useState(false);
  const [currentPage, setCurrentPage] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [selectedMod, setSelectedMod] = useState<CurseForgeMod | null>(null);
  const [isLoadingDetails, setIsLoadingDetails] = useState(false);

  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const loadInstalledMods = useCallback(async () => {
    try {
      const mods = await getInstalledMods();
      setInstalledMods(mods || []);
    } catch (err) {
      console.error('Failed to load installed mods:', err);
    }
  }, [getInstalledMods]);

  const loadCategories = useCallback(async () => {
    try {
      const cats = await getModCategories();
      setCategories(cats || []);
    } catch (err) {
      console.error('Failed to load categories:', err);
    }
  }, [getModCategories]);

  const handleCheckUpdates = useCallback(async () => {
    if (!checkModUpdates) return;
    
    setIsCheckingUpdates(true);
    try {
      const updates = await checkModUpdates();
      const updateMap: Record<string, { hasUpdate: boolean; newVersion: string }> = {};
      updates.forEach(u => {
        updateMap[u.modId] = { hasUpdate: u.hasUpdate, newVersion: u.newVersion };
      });
      setModUpdates(updateMap);
    } catch (err) {
      console.error('Failed to check updates:', err);
    }
    setIsCheckingUpdates(false);
  }, [checkModUpdates]);

  useEffect(() => {
    loadInstalledMods();
    loadCategories();
  }, [loadInstalledMods, loadCategories]);

  // Auto-load popular mods on browse tab open
  const loadPopularMods = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setCurrentPage(0);
    try {
      console.log('[ModManager] Loading mods...');
      const result = await searchMods('', 0, 0);
      console.log('[ModManager] Result:', result);
      setSearchResults(result?.mods || []);
      setHasMore((result?.mods?.length || 0) >= 20);
      if (!result?.mods?.length) {
        console.log('[ModManager] No mods returned from API');
      }
    } catch (err: any) {
      console.error('[ModManager] Error loading mods:', err);
      setError(err.message || 'Failed to load mods');
      setSearchResults([]);
    }
    setIsLoading(false);
  }, [searchMods]);

  // Load more mods for infinite scroll
  const loadMoreMods = useCallback(async () => {
    if (isLoadingMore || !hasMore) return;
    
    setIsLoadingMore(true);
    const nextPage = currentPage + 1;
    
    try {
      const result = await searchMods(searchQuery, selectedCategory, nextPage);
      const newMods = result?.mods || [];
      
      if (newMods.length > 0) {
        setSearchResults(prev => [...prev, ...newMods]);
        setCurrentPage(nextPage);
        setHasMore(newMods.length >= 20);
      } else {
        setHasMore(false);
      }
    } catch (err: any) {
      console.error('[ModManager] Error loading more:', err);
    }
    setIsLoadingMore(false);
  }, [searchMods, searchQuery, selectedCategory, currentPage, isLoadingMore, hasMore]);

  // Handle scroll for infinite loading
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const target = e.target as HTMLDivElement;
    const scrollBottom = target.scrollHeight - target.scrollTop - target.clientHeight;
    
    if (scrollBottom < 200 && !isLoadingMore && hasMore && activeTab === 'browse') {
      loadMoreMods();
    }
  }, [loadMoreMods, isLoadingMore, hasMore, activeTab]);

  useEffect(() => {
    if (activeTab === 'browse' && searchResults.length === 0) {
      loadPopularMods();
    }
  }, [activeTab, searchResults.length, loadPopularMods]);

  // Re-sort when sortBy changes
  useEffect(() => {
    if (searchResults.length === 0) return;
    
    const sortedMods = [...searchResults];
    if (sortBy === 'downloads') {
      sortedMods.sort((a, b) => b.downloadCount - a.downloadCount);
    } else if (sortBy === 'updated') {
      sortedMods.sort((a, b) => new Date(b.dateModified || 0).getTime() - new Date(a.dateModified || 0).getTime());
    } else if (sortBy === 'name') {
      sortedMods.sort((a, b) => a.name.localeCompare(b.name));
    }
    setSearchResults(sortedMods);
  }, [sortBy]);

  const handleSearch = async () => {
    setIsLoading(true);
    setError(null);
    setCurrentPage(0);
    try {
      const result = await searchMods(searchQuery, selectedCategory, 0);
      let mods = result.mods || [];
      
      // Sort results
      if (sortBy === 'downloads') {
        mods.sort((a, b) => b.downloadCount - a.downloadCount);
      } else if (sortBy === 'updated') {
        mods.sort((a, b) => new Date(b.dateModified || 0).getTime() - new Date(a.dateModified || 0).getTime());
      } else if (sortBy === 'name') {
        mods.sort((a, b) => a.name.localeCompare(b.name));
      }
      
      setSearchResults(mods);
      setHasMore(mods.length >= 20);
    } catch (err: any) {
      setError(err.message || 'Failed to search mods');
      setSearchResults([]);
    }
    setIsLoading(false);
  };

  const handleInstall = async (mod: CurseForgeMod) => {
    setIsInstalling(mod.id);
    try {
      await installMod(mod.id);
      await loadInstalledMods();
    } catch (err: any) {
      setError(err.message || 'Failed to install mod');
    }
    setIsInstalling(null);
  };

  const handleViewDetails = async (mod: CurseForgeMod) => {
    setIsLoadingDetails(true);
    setSelectedMod(mod);
    
    // If we have getModDetails, fetch full details with screenshots
    if (getModDetails) {
      try {
        const details = await getModDetails(mod.id);
        setSelectedMod(details);
      } catch (err) {
        console.error('Failed to load mod details:', err);
        // Keep the basic mod info if details fail
      }
    }
    setIsLoadingDetails(false);
  };

  const handleUninstall = async (modId: string) => {
    try {
      await uninstallMod(modId);
      await loadInstalledMods();
    } catch (err: any) {
      setError(err.message || 'Failed to uninstall mod');
    }
  };

  const handleToggle = async (modId: string, enabled: boolean) => {
    try {
      await toggleMod(modId, enabled);
      await loadInstalledMods();
    } catch (err: any) {
      setError(err.message || 'Failed to toggle mod');
    }
  };

  const isModInstalled = (cfModId: number) => {
    return installedMods.some(m => m.id === `cf-${cfModId}`);
  };

  // Count mods with actual updates available
  const updateCount = Object.values(modUpdates).filter(u => u.hasUpdate).length;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-8"
    >
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        className="w-full max-w-4xl max-h-[85vh] bg-[#0d0d0d] rounded-2xl border border-white/10 overflow-hidden flex flex-col"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-white/10">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
              <Package size={20} className="text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">Mod Manager</h2>
              <p className="text-xs text-gray-400">Browse and manage Hytale mods</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={openModsFolder}
              className="p-2 rounded-lg hover:bg-white/5 text-gray-400 hover:text-white transition-colors"
              title="Open Mods Folder"
            >
              <FolderOpen size={18} />
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-white/5 text-gray-400 hover:text-white transition-colors"
            >
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-white/10">
          <button
            onClick={() => setActiveTab('installed')}
            className={`flex-1 py-3 text-sm font-medium transition-colors relative ${
              activeTab === 'installed'
                ? 'text-[#FFA845] border-b-2 border-[#FFA845]'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            Installed ({installedMods.length})
            {updateCount > 0 && (
              <span className="absolute top-2 ml-1 px-1.5 py-0.5 text-xs bg-green-500 text-white rounded-full">
                {updateCount}
              </span>
            )}
          </button>
          <button
            onClick={() => setActiveTab('browse')}
            className={`flex-1 py-3 text-sm font-medium transition-colors ${
              activeTab === 'browse'
                ? 'text-[#FFA845] border-b-2 border-[#FFA845]'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            Browse CurseForge
          </button>
        </div>

        {/* Search Bar (Browse Tab) */}
        {activeTab === 'browse' && (
          <div className="p-4 border-b border-white/5 space-y-3">
            <div className="flex gap-2">
              <div className="flex-1 relative">
                <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    // Search as you type with debounce
                    const query = e.target.value;
                    setTimeout(() => {
                      if (query === e.target.value) {
                        handleSearch();
                      }
                    }, 300);
                  }}
                  placeholder="Search mods..."
                  className="w-full bg-white/5 border border-white/10 rounded-lg pl-10 pr-4 py-2 text-white placeholder-gray-500 focus:outline-none focus:border-[#FFA845]/50"
                />
              </div>
              <select
                value={selectedCategory}
                onChange={(e) => {
                  setSelectedCategory(Number(e.target.value));
                  setTimeout(handleSearch, 100);
                }}
                className="bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-[#FFA845]/50"
              >
                <option value={0}>All Categories</option>
                {categories.map((cat) => (
                  <option key={cat.id} value={cat.id}>{cat.name}</option>
                ))}
              </select>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
                className="bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-[#FFA845]/50"
              >
                <option value="downloads">Most Downloads</option>
                <option value="updated">Recently Updated</option>
                <option value="name">Name (A-Z)</option>
              </select>
            </div>
            {searchResults.length === 0 && !isLoading && (
              <div className="flex gap-2">
                <button
                  onClick={loadPopularMods}
                  className="flex items-center gap-2 px-3 py-1.5 bg-white/5 hover:bg-white/10 rounded-lg text-sm text-gray-300 transition-colors"
                >
                  <TrendingUp size={14} />
                  Load Popular Mods
                </button>
              </div>
            )}
          </div>
        )}

        {/* Installed Tab Actions */}
        {activeTab === 'installed' && installedMods.length > 0 && (
          <div className="p-4 border-b border-white/5">
            <div className="flex gap-2">
              <motion.button
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={handleCheckUpdates}
                disabled={isCheckingUpdates}
                className="flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm text-gray-300 transition-colors disabled:opacity-50"
              >
                {isCheckingUpdates ? (
                  <Loader2 size={16} className="animate-spin" />
                ) : (
                  <RefreshCw size={16} />
                )}
                Check for Updates
              </motion.button>
              {updateCount > 0 && (
                <span className="flex items-center gap-1 px-3 py-2 bg-green-500/10 border border-green-500/20 rounded-lg text-sm text-green-400">
                  <ArrowUp size={14} />
                  {updateCount} update{updateCount !== 1 ? 's' : ''} available
                </span>
              )}
            </div>
          </div>
        )}

        {/* Error Message */}
        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="mx-4 mt-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg flex items-center gap-2"
            >
              <AlertCircle size={16} className="text-red-400" />
              <span className="text-sm text-red-400">{error}</span>
              <button onClick={() => setError(null)} className="ml-auto text-red-400 hover:text-red-300">
                <X size={14} />
              </button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Content */}
        <div 
          ref={scrollContainerRef}
          onScroll={handleScroll}
          className="flex-1 overflow-y-auto p-4"
          style={{ minHeight: '400px' }}
        >
          {activeTab === 'installed' ? (
            installedMods.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                <Package size={48} className="mb-4 opacity-50" />
                <p className="text-lg font-medium">No mods installed</p>
                <p className="text-sm">Browse CurseForge to find and install mods</p>
              </div>
            ) : (
              <div className="space-y-3">
                {installedMods.map((mod) => (
                  <motion.div
                    key={mod.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className={`p-4 bg-white/5 rounded-xl border ${modUpdates[mod.id]?.hasUpdate ? 'border-green-500/30' : 'border-white/5'} ${!mod.enabled ? 'opacity-50' : ''}`}
                  >
                    <div className="flex items-start gap-4">
                      {mod.iconUrl ? (
                        <img src={mod.iconUrl} alt={mod.name} className="w-12 h-12 rounded-lg" />
                      ) : (
                        <div className="w-12 h-12 rounded-lg bg-white/10 flex items-center justify-center">
                          <Package size={20} className="text-gray-400" />
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="font-bold text-white truncate">{mod.name}</h3>
                          <span className="text-xs text-gray-400 bg-white/5 px-2 py-0.5 rounded">
                            {mod.version}
                          </span>
                          {modUpdates[mod.id]?.hasUpdate && (
                            <span className="flex items-center gap-1 text-xs text-green-400 bg-green-500/10 px-2 py-0.5 rounded">
                              <ArrowUp size={10} />
                              {modUpdates[mod.id].newVersion}
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-400">by {mod.author}</p>
                        <p className="text-xs text-gray-500 mt-1 line-clamp-2">{mod.description}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        {modUpdates[mod.id]?.hasUpdate && (
                          <button
                            onClick={() => {
                              // Re-install the mod with the new version
                              if (mod.curseForgeId) {
                                // Trigger reinstall with CurseForge ID
                                setActiveTab('browse');
                                setSearchQuery(mod.name);
                                handleSearch();
                              }
                            }}
                            className="px-3 py-1.5 rounded-lg bg-green-500/20 hover:bg-green-500/30 text-green-400 text-sm font-medium transition-colors"
                            title="Update mod"
                          >
                            Update
                          </button>
                        )}
                        <button
                          onClick={() => handleToggle(mod.id, !mod.enabled)}
                          className="p-2 rounded-lg hover:bg-white/10 transition-colors"
                          title={mod.enabled ? 'Disable' : 'Enable'}
                        >
                          {mod.enabled ? (
                            <ToggleRight size={24} className="text-green-400" />
                          ) : (
                            <ToggleLeft size={24} className="text-gray-400" />
                          )}
                        </button>
                        <button
                          onClick={() => handleUninstall(mod.id)}
                          className="p-2 rounded-lg hover:bg-red-500/20 text-red-400 transition-colors"
                          title="Uninstall"
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </div>
                  </motion.div>
                ))}
              </div>
            )
          ) : (
            // Browse Tab
            isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 size={32} className="text-[#FFA845] animate-spin" />
              </div>
            ) : searchResults.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                <Package size={48} className="mb-4 opacity-50" />
                <p className="text-lg font-medium">No Mods Available</p>
                <p className="text-sm text-center max-w-md">
                  Hytale is currently in early access. Mod support will be available when the game officially releases.
                  <br />
                  You can still manage locally installed mods in the Installed tab.
                </p>
                <button
                  onClick={loadPopularMods}
                  className="mt-4 flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm transition-colors"
                >
                  <RefreshCw size={16} />
                  Retry
                </button>
              </div>
            ) : (
              <div className="grid gap-3">
                {searchResults.map((mod) => {
                  const installed = isModInstalled(mod.id);
                  return (
                    <motion.div
                      key={mod.id}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      onClick={() => handleViewDetails(mod)}
                      className="p-4 bg-white/5 rounded-xl border border-white/5 hover:border-[#FFA845]/30 hover:bg-white/[0.07] transition-colors cursor-pointer"
                    >
                      <div className="flex items-start gap-4">
                        {mod.logo?.thumbnailUrl ? (
                          <img src={mod.logo.thumbnailUrl} alt={mod.name} className="w-12 h-12 rounded-lg" />
                        ) : (
                          <div className="w-12 h-12 rounded-lg bg-white/10 flex items-center justify-center">
                            <Package size={20} className="text-gray-400" />
                          </div>
                        )}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <h3 className="font-bold text-white truncate">{mod.name}</h3>
                            <a
                              href={`https://www.curseforge.com/hytale/${mod.slug}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-gray-400 hover:text-white"
                              onClick={(e) => e.stopPropagation()}
                            >
                              <ExternalLink size={14} />
                            </a>
                          </div>
                          <p className="text-sm text-gray-400">
                            by {mod.authors?.[0]?.name || 'Unknown'}
                            <span className="mx-2">•</span>
                            {formatDownloads(mod.downloadCount)} downloads
                            {mod.dateModified && (
                              <>
                                <span className="mx-2">•</span>
                                <Clock size={12} className="inline mr-1" />
                                {formatDate(mod.dateModified)}
                              </>
                            )}
                          </p>
                          <p className="text-xs text-gray-500 mt-1 line-clamp-2">{mod.summary}</p>
                        </div>
                        <motion.button
                          whileHover={{ scale: 1.05 }}
                          whileTap={{ scale: 0.95 }}
                          onClick={(e) => {
                            e.stopPropagation();
                            !installed && handleInstall(mod);
                          }}
                          disabled={installed || isInstalling === mod.id}
                          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                            installed
                              ? 'bg-green-500/20 text-green-400 cursor-default'
                              : 'bg-[#FFA845] text-black hover:bg-[#FFB865]'
                          } disabled:opacity-50`}
                        >
                          {isInstalling === mod.id ? (
                            <Loader2 size={18} className="animate-spin" />
                          ) : installed ? (
                            'Installed'
                          ) : (
                            <Download size={18} />
                          )}
                        </motion.button>
                      </div>
                    </motion.div>
                  );
                })}
                
                {/* Infinite scroll loading indicator */}
                {isLoadingMore && (
                  <div className="flex justify-center py-4">
                    <Loader2 size={24} className="text-[#FFA845] animate-spin" />
                  </div>
                )}
                
                {!hasMore && searchResults.length > 0 && (
                  <div className="text-center py-4 text-gray-500 text-sm">
                    Showing all {searchResults.length} mods
                  </div>
                )}
              </div>
            )
          )}
        </div>
      </motion.div>

      {/* Mod Detail Modal */}
      <AnimatePresence>
        {selectedMod && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setSelectedMod(null)}
            className="absolute inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-8"
          >
            <motion.div
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              onClick={(e) => e.stopPropagation()}
              className="bg-[#1a1a1a] rounded-2xl border border-white/10 w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col"
            >
              {/* Header */}
              <div className="p-6 border-b border-white/10 flex items-start gap-4">
                {selectedMod.logo?.url || selectedMod.logo?.thumbnailUrl ? (
                  <img 
                    src={selectedMod.logo?.url || selectedMod.logo?.thumbnailUrl} 
                    alt={selectedMod.name} 
                    className="w-20 h-20 rounded-xl object-cover"
                  />
                ) : (
                  <div className="w-20 h-20 rounded-xl bg-white/10 flex items-center justify-center">
                    <Package size={32} className="text-gray-400" />
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <h2 className="text-2xl font-bold text-white">{selectedMod.name}</h2>
                  <p className="text-gray-400 mt-1">
                    by {selectedMod.authors?.[0]?.name || 'Unknown'}
                    <span className="mx-2">•</span>
                    {formatDownloads(selectedMod.downloadCount)} downloads
                  </p>
                  {selectedMod.categories?.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-2">
                      {selectedMod.categories.slice(0, 3).map((cat, i) => (
                        <span key={i} className="px-2 py-0.5 bg-white/10 rounded text-xs text-gray-300">
                          {cat.name}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
                <button
                  onClick={() => setSelectedMod(null)}
                  className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                >
                  <X size={20} />
                </button>
              </div>

              {/* Content */}
              <div className="flex-1 overflow-y-auto p-6">
                {isLoadingDetails ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 size={32} className="text-[#FFA845] animate-spin" />
                  </div>
                ) : (
                  <>
                    {/* Description */}
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-white mb-2">Description</h3>
                      <p className="text-gray-300 text-sm leading-relaxed">{selectedMod.summary}</p>
                    </div>

                    {/* Screenshots */}
                    {selectedMod.screenshots && selectedMod.screenshots.length > 0 && (
                      <div>
                        <h3 className="text-lg font-semibold text-white mb-3">Screenshots</h3>
                        <div className="grid grid-cols-2 gap-3">
                          {selectedMod.screenshots.slice(0, 4).map((screenshot) => (
                            <a
                              key={screenshot.id}
                              href={screenshot.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="block aspect-video rounded-lg overflow-hidden border border-white/10 hover:border-[#FFA845]/50 transition-colors"
                            >
                              <img
                                src={screenshot.thumbnailUrl || screenshot.url}
                                alt={screenshot.title || 'Screenshot'}
                                className="w-full h-full object-cover"
                              />
                            </a>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Mod Info */}
                    <div className="mt-6 pt-4 border-t border-white/10">
                      <div className="flex items-center gap-4 text-sm text-gray-400">
                        {selectedMod.dateModified && (
                          <span>
                            <Clock size={14} className="inline mr-1" />
                            Updated {formatDate(selectedMod.dateModified)}
                          </span>
                        )}
                        <a
                          href={`https://www.curseforge.com/hytale/${selectedMod.slug}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 text-[#FFA845] hover:underline"
                        >
                          <ExternalLink size={14} />
                          View on CurseForge
                        </a>
                      </div>
                    </div>
                  </>
                )}
              </div>

              {/* Footer */}
              <div className="p-6 border-t border-white/10">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => {
                    handleInstall(selectedMod);
                    setSelectedMod(null);
                  }}
                  disabled={isModInstalled(selectedMod.id) || isInstalling === selectedMod.id}
                  className={`w-full py-3 rounded-xl font-bold text-lg transition-colors ${
                    isModInstalled(selectedMod.id)
                      ? 'bg-green-500/20 text-green-400 cursor-default'
                      : 'bg-gradient-to-r from-[#FFA845] to-[#FF6B35] text-black hover:shadow-lg'
                  }`}
                >
                  {isInstalling === selectedMod.id ? (
                    <Loader2 size={24} className="animate-spin mx-auto" />
                  ) : isModInstalled(selectedMod.id) ? (
                    'Already Installed'
                  ) : (
                    <span className="flex items-center justify-center gap-2">
                      <Download size={20} />
                      Install Mod
                    </span>
                  )}
                </motion.button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
};
