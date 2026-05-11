import { Upload } from 'lucide-react';
import { useRef } from 'react';

interface UploadZoneProps {
  onFileUpload: (files: FileList) => void;
  disabled?: boolean;
}

export function UploadZone({ onFileUpload, disabled = false }: UploadZoneProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (disabled) return;
    if (e.dataTransfer.files.length > 0) {
      onFileUpload(e.dataTransfer.files);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleClick = () => {
    if (disabled) return;
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      onFileUpload(e.target.files);
    }
  };

  return (
    <div
      onClick={handleClick}
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      className="border-2 border-dashed border-blue-300 rounded-2xl p-12 text-center cursor-pointer hover:border-blue-500 hover:bg-blue-50/50 transition-all"
    >
      <input
        ref={fileInputRef}
        type="file"
        multiple
        disabled={disabled}
        onChange={handleFileChange}
        className="hidden"
      />
      <Upload className="w-16 h-16 mx-auto mb-4 text-blue-500" />
      <h3 className="mb-2">Upload Files</h3>
      <p className="text-muted-foreground">
        {disabled ? 'Uploading...' : 'Click or drag files here to upload'}
      </p>
    </div>
  );
}
