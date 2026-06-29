import { useState, useCallback, useRef } from 'react'
import { useDropzone } from 'react-dropzone'
import { validateFile } from '../utils/fileValidation'

type UploadStatus = 'idle' | 'validating' | 'uploading' | 'success' | 'error'

function Upload() {
  const [status, setStatus] = useState<UploadStatus>('idle')
  const [progress, setProgress] = useState(0)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [fileName, setFileName] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const cameraInputRef = useRef<HTMLInputElement>(null)

  const handleFile = useCallback(async (file: File) => {
    setStatus('validating')
    setErrorMessage(null)
    setFileName(file.name)

    const validation = validateFile(file)
    if (!validation.valid) {
      setStatus('error')
      setErrorMessage(validation.error ?? 'File validation failed.')
      return
    }

    // Simulate presigned URL fetch and S3 upload
    setStatus('uploading')
    setProgress(0)

    // Simulate upload progress
    const interval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 100) {
          clearInterval(interval)
          return 100
        }
        return prev + 20
      })
    }, 300)

    // Simulate upload completion
    setTimeout(() => {
      clearInterval(interval)
      setProgress(100)
      setStatus('success')
    }, 1800)
  }, [])

  const onDrop = useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles.length > 0) {
      handleFile(acceptedFiles[0])
    }
  }, [handleFile])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: false,
    noClick: false,
  })

  const handleCameraCapture = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files
    if (files && files.length > 0) {
      handleFile(files[0])
    }
  }

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files
    if (files && files.length > 0) {
      handleFile(files[0])
    }
  }

  const resetUpload = () => {
    setStatus('idle')
    setProgress(0)
    setErrorMessage(null)
    setFileName(null)
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8">
      <h1 className="text-2xl font-bold text-gray-900">Document Upload</h1>
      <p className="mt-2 text-gray-600">Upload payment documents for processing</p>

      <div className="mt-6 max-w-2xl">
        {/* Upload status display */}
        {status === 'validating' && (
          <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-blue-700 font-medium">Validating file...</p>
          </div>
        )}

        {status === 'uploading' && (
          <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-blue-700 font-medium">
              Uploading {fileName}...
            </p>
            <div className="mt-2 w-full bg-blue-200 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${progress}%` }}
              />
            </div>
            <p className="mt-1 text-sm text-blue-600">{progress}%</p>
          </div>
        )}

        {status === 'success' && (
          <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
            <p className="text-green-700 font-medium">
              ✓ {fileName} uploaded successfully
            </p>
            <p className="mt-1 text-sm text-green-600">
              Document is being processed by the payment pipeline.
            </p>
            <button
              onClick={resetUpload}
              className="mt-3 text-sm text-green-700 underline hover:text-green-900"
            >
              Upload another document
            </button>
          </div>
        )}

        {status === 'error' && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-700 font-medium">Upload failed</p>
            <p className="mt-1 text-sm text-red-600">{errorMessage}</p>
            <button
              onClick={resetUpload}
              className="mt-3 text-sm text-red-700 underline hover:text-red-900"
            >
              Try again
            </button>
          </div>
        )}

        {/* Mobile upload interface: camera capture primary, file picker secondary */}
        {status === 'idle' && (
          <div className="sm:hidden space-y-4">
            <button
              onClick={() => cameraInputRef.current?.click()}
              className="w-full flex items-center justify-center gap-2 px-6 py-4 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Capture Document
            </button>
            <input
              ref={cameraInputRef}
              type="file"
              accept="image/*"
              capture="environment"
              onChange={handleCameraCapture}
              className="hidden"
              aria-label="Camera capture"
            />

            <button
              onClick={() => fileInputRef.current?.click()}
              className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-white text-gray-700 font-medium rounded-lg border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Choose from files
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf,.png,.jpeg,.jpg,.tiff,.tif"
              onChange={handleFileSelect}
              className="hidden"
              aria-label="File picker"
            />

            <p className="text-center text-xs text-gray-500">
              Supported: PDF, PNG, JPEG, TIFF (max 10MB)
            </p>
          </div>
        )}

        {/* Desktop/tablet drag-and-drop zone */}
        {status === 'idle' && (
          <div className="hidden sm:block">
            <div
              {...getRootProps()}
              className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
                isDragActive
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-300 hover:border-gray-400 hover:bg-gray-50'
              }`}
            >
              <input {...getInputProps()} aria-label="Drop zone file input" />
              <svg className="mx-auto w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              {isDragActive ? (
                <p className="mt-4 text-blue-600 font-medium">Drop the file here</p>
              ) : (
                <>
                  <p className="mt-4 text-gray-700 font-medium">
                    Drag and drop a document here, or click to select
                  </p>
                  <p className="mt-2 text-sm text-gray-500">
                    Supported: PDF, PNG, JPEG, TIFF (max 10MB)
                  </p>
                </>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default Upload
