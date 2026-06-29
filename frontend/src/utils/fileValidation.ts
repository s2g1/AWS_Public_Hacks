export interface FileValidationResult {
  valid: boolean
  error?: string
}

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

const SUPPORTED_MIME_TYPES = [
  'application/pdf',
  'image/png',
  'image/jpeg',
  'image/tiff',
]

const SUPPORTED_EXTENSIONS = [
  '.pdf',
  '.png',
  '.jpeg',
  '.jpg',
  '.tiff',
  '.tif',
]

function getFileExtension(filename: string): string {
  const lastDot = filename.lastIndexOf('.')
  if (lastDot === -1) return ''
  return filename.slice(lastDot).toLowerCase()
}

export function validateFile(file: File): FileValidationResult {
  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    return {
      valid: false,
      error: `File size exceeds 10MB limit. File is ${(file.size / (1024 * 1024)).toFixed(1)}MB.`,
    }
  }

  // Check file format by MIME type and extension
  const extension = getFileExtension(file.name)
  const hasValidMime = SUPPORTED_MIME_TYPES.includes(file.type)
  const hasValidExtension = SUPPORTED_EXTENSIONS.includes(extension)

  if (!hasValidMime && !hasValidExtension) {
    return {
      valid: false,
      error: `Unsupported file format. Supported formats: PDF, PNG, JPEG, TIFF.`,
    }
  }

  return { valid: true }
}
