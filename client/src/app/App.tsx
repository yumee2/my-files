import { useState, useEffect } from 'react';
import { Search, Cloud, HardDrive } from 'lucide-react';
import { FileCard } from './components/FileCard';
import { UploadZone } from './components/UploadZone';
import {
  ApiError,
  deleteFile,
  downloadFile,
  getFiles,
  login,
  StoredFile,
  uploadFile,
} from './lib/api';

export default function App() {
  const [files, setFiles] = useState<StoredFile[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [storageUsed, setStorageUsed] = useState(0);
  const [password, setPassword] = useState('');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadFiles();
  }, []);

  const calculateStorage = (fileList: StoredFile[]) => {
    const total = fileList.reduce((acc, file) => acc + file.size, 0);
    setStorageUsed(total);
  };

  const loadFiles = async () => {
    setIsLoading(true);
    setError('');

    try {
      const serverFiles = await getFiles();
      setFiles(serverFiles);
      calculateStorage(serverFiles);
      setIsAuthenticated(true);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setIsAuthenticated(false);
        setFiles([]);
        calculateStorage([]);
      } else {
        setError(err instanceof Error ? err.message : 'Failed to load files');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      await login(password);
      setPassword('');
      await loadFiles();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileUpload = async (fileList: FileList) => {
    setIsUploading(true);
    setError('');

    try {
      await Promise.all(Array.from(fileList).map((file) => uploadFile(file)));
      await loadFiles();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setIsUploading(false);
    }
  };

  const handleDownload = async (file: StoredFile) => {
    setError('');

    try {
      await downloadFile(file);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Download failed');
    }
  };

  const handleDelete = async (id: string) => {
    setError('');

    try {
      await deleteFile(id);
      const updatedFiles = files.filter(f => f.id !== id);
      setFiles(updatedFiles);
      calculateStorage(updatedFiles);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  const formatStorageSize = (bytes: number) => {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const filteredFiles = files.filter(file =>
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (!isAuthenticated && !isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <form onSubmit={handleLogin} className="w-full max-w-sm bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
          <div className="flex items-center gap-3 mb-6">
            <Cloud className="w-8 h-8 text-blue-500" />
            <h1>My Files</h1>
          </div>
          <label htmlFor="password" className="block mb-2 text-gray-700">
            Password
          </label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 bg-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            autoComplete="current-password"
            required
          />
          {error && <p className="text-red-600 mt-3">{error}</p>}
          <button
            type="submit"
            className="w-full mt-6 py-3 rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-colors"
          >
            Sign in
          </button>
        </form>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <Cloud className="w-8 h-8 text-blue-500" />
              <h1>My Files</h1>
            </div>
            <div className="flex items-center gap-2 text-muted-foreground">
              <HardDrive className="w-5 h-5" />
              <span>{formatStorageSize(storageUsed)} used</span>
            </div>
          </div>

          {/* Search Bar */}
          <div className="relative">
            <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search files..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-12 pr-4 py-3 bg-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <UploadZone onFileUpload={handleFileUpload} disabled={isUploading} />
        </div>

        {error && (
          <div className="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700">
            {error}
          </div>
        )}

        {isLoading ? (
          <div className="text-center py-16 text-muted-foreground">Loading files...</div>
        ) : filteredFiles.length === 0 ? (
          <div className="text-center py-16">
            <Cloud className="w-20 h-20 mx-auto mb-4 text-gray-300" />
            <h2 className="mb-2 text-gray-600">
              {searchQuery ? 'No files found' : 'No files yet'}
            </h2>
            <p className="text-muted-foreground">
              {searchQuery ? 'Try a different search term' : 'Upload your first file to get started'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredFiles.map((file) => (
              <FileCard
                key={file.id}
                file={file}
                onDownload={handleDownload}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
