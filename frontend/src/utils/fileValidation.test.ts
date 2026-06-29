import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import { validateFile } from './fileValidation'

/**
 * Property 26: File Upload Validation
 * Validates: Requirement 24.5
 *
 * For any uploaded file, the system rejects files exceeding 10MB or with
 * unsupported formats (not PDF, PNG, JPEG, or TIFF) before initiating upload to S3.
 */

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

const SUPPORTED_EXTENSIONS = ['.pdf', '.png', '.jpeg', '.jpg', '.tiff', '.tif']
const SUPPORTED_MIME_TYPES = [
  'application/pdf',
  'image/png',
  'image/jpeg',
  'image/tiff',
]

// Helper to create a File object with a specific size
function createFileWithSize(
  size: number,
  name: string,
  mimeType: string
): File {
  const content = new Uint8Array(size)
  return new File([content], name, { type: mimeType })
}

// Arbitrary for supported file formats (extension + mime type pairs)
const supportedFormatArb = fc.oneof(
  fc.constant({ ext: '.pdf', mime: 'application/pdf' }),
  fc.constant({ ext: '.png', mime: 'image/png' }),
  fc.constant({ ext: '.jpeg', mime: 'image/jpeg' }),
  fc.constant({ ext: '.jpg', mime: 'image/jpeg' }),
  fc.constant({ ext: '.tiff', mime: 'image/tiff' }),
  fc.constant({ ext: '.tif', mime: 'image/tiff' })
)

// Arbitrary for unsupported file formats
const unsupportedFormatArb = fc
  .tuple(
    fc.string({ minLength: 1, maxLength: 5, unit: fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz') }),
    fc.string({ minLength: 1, maxLength: 10, unit: fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz') })
  )
  .filter(([ext, mime]) => {
    const dotExt = `.${ext}`
    return (
      !SUPPORTED_EXTENSIONS.includes(dotExt) &&
      !SUPPORTED_MIME_TYPES.includes(`application/${mime}`) &&
      !SUPPORTED_MIME_TYPES.includes(`image/${mime}`)
    )
  })
  .map(([ext, mime]) => ({ ext: `.${ext}`, mime: `application/${mime}` }))

// Arbitrary for file sizes > 10MB (oversized)
const oversizedFileArb = fc.integer({ min: MAX_FILE_SIZE + 1, max: 50 * 1024 * 1024 })

// Arbitrary for file sizes <= 10MB (valid size)
const validSizeArb = fc.integer({ min: 1, max: MAX_FILE_SIZE })

// Arbitrary for base filenames without extension
const baseFilenameArb = fc.string({
  minLength: 1,
  maxLength: 20,
  unit: fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789_-'),
})

describe('Property 26: File Upload Validation', () => {
  /**
   * **Validates: Requirement 24.5**
   * Property: Any file > 10MB always returns {valid: false} with a size error
   */
  it('rejects any file exceeding 10MB with a size error', () => {
    fc.assert(
      fc.property(
        oversizedFileArb,
        supportedFormatArb,
        baseFilenameArb,
        (size, format, baseName) => {
          const file = createFileWithSize(
            size,
            `${baseName}${format.ext}`,
            format.mime
          )
          const result = validateFile(file)

          expect(result.valid).toBe(false)
          expect(result.error).toBeDefined()
          expect(result.error!.toLowerCase()).toContain('size')
        }
      ),
      { numRuns: 100 }
    )
  })

  /**
   * **Validates: Requirement 24.5**
   * Property: Any file <= 10MB with a supported format always returns {valid: true}
   */
  it('accepts any file <= 10MB with a supported format', () => {
    fc.assert(
      fc.property(
        validSizeArb,
        supportedFormatArb,
        baseFilenameArb,
        (size, format, baseName) => {
          const file = createFileWithSize(
            size,
            `${baseName}${format.ext}`,
            format.mime
          )
          const result = validateFile(file)

          expect(result.valid).toBe(true)
          expect(result.error).toBeUndefined()
        }
      ),
      { numRuns: 100 }
    )
  })

  /**
   * **Validates: Requirement 24.5**
   * Property: Any file with an unsupported format always returns {valid: false} with a format error
   */
  it('rejects any file with an unsupported format', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: MAX_FILE_SIZE }),
        unsupportedFormatArb,
        baseFilenameArb,
        (size, format, baseName) => {
          const file = createFileWithSize(
            size,
            `${baseName}${format.ext}`,
            format.mime
          )
          const result = validateFile(file)

          expect(result.valid).toBe(false)
          expect(result.error).toBeDefined()
          expect(result.error!.toLowerCase()).toContain('format')
        }
      ),
      { numRuns: 100 }
    )
  })

  /**
   * **Validates: Requirement 24.5**
   * Property: Empty/zero-byte supported files are valid (size check passes)
   */
  it('accepts empty/zero-byte files with supported formats', () => {
    fc.assert(
      fc.property(supportedFormatArb, baseFilenameArb, (format, baseName) => {
        const file = createFileWithSize(
          0,
          `${baseName}${format.ext}`,
          format.mime
        )
        const result = validateFile(file)

        expect(result.valid).toBe(true)
        expect(result.error).toBeUndefined()
      }),
      { numRuns: 50 }
    )
  })
})
