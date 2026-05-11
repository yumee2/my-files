import { FileText, Image, Video, Music, File, Download, Trash2 } from 'lucide-react';

interface FileCardProps {
  file: {
    id: string;
    name: string;
    size: number;
    type: string;
    uploadedAt?: number;
  };
  onDownload: (file: FileCardProps['file']) => void;
  onDelete: (id: string) => void;
}

export function FileCard({ file, onDownload, onDelete }: FileCardProps) {
  const getFileIcon = () => {
    if (file.type.startsWith('image/')) return <Image className="w-12 h-12 text-blue-500" />;
    if (file.type.startsWith('video/')) return <Video className="w-12 h-12 text-purple-500" />;
    if (file.type.startsWith('audio/')) return <Music className="w-12 h-12 text-pink-500" />;
    if (file.type.includes('pdf') || file.type.includes('document')) return <FileText className="w-12 h-12 text-orange-500" />;
    return <File className="w-12 h-12 text-gray-500" />;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const formatDate = (timestamp?: number) => {
    if (!timestamp) return 'Stored on server';
    const date = new Date(timestamp);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  return (
    <div className="bg-white rounded-xl p-6 shadow-sm hover:shadow-md transition-shadow border border-gray-100 flex flex-col items-center">
      <div className="mb-4">
        {getFileIcon()}
      </div>
      <h3 className="text-center mb-2 w-full truncate px-2">{file.name}</h3>
      <p className="text-muted-foreground mb-1">{formatFileSize(file.size)}</p>
      <p className="text-muted-foreground mb-4">{formatDate(file.uploadedAt)}</p>
      <div className="flex gap-2 mt-auto">
        <button
          onClick={() => onDownload(file)}
          className="p-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-colors"
          aria-label="Download file"
        >
          <Download className="w-4 h-4" />
        </button>
        <button
          onClick={() => onDelete(file.id)}
          className="p-2 rounded-lg bg-red-500 text-white hover:bg-red-600 transition-colors"
          aria-label="Delete file"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
