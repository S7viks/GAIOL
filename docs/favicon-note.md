# Favicon Setup Note

## Important: Create Favicon Files

The HTML files now reference favicon files, but you need to create them:

### Required Files:
1. `/web/favicon.ico` - Standard favicon (16x16, 32x32, 48x48)
2. `/web/favicon-32x32.png` - 32x32 PNG favicon
3. `/web/favicon-192x192.png` - 192x192 PNG favicon (for PWA/app icons)

### Quick Options:

#### Option 1: Use an Online Generator
1. Go to [favicon.io](https://favicon.io) or [realfavicongenerator.net](https://realfavicongenerator.net)
2. Upload a logo/image or generate from text
3. Download the generated files
4. Place them in the `/web/` directory

#### Option 2: Create Simple Text Favicon
1. Use [favicon.io text generator](https://favicon.io/favicon-generator/)
2. Enter "G" or "GAIOL" as text
3. Choose colors matching your theme (cyan/teal)
4. Download and place files in `/web/`

#### Option 3: Use a Simple SVG
Create a simple SVG logo and convert it to favicon formats.

### Recommended Design:
- Use your accent color (#00e5ff) as primary
- Keep it simple and recognizable at small sizes
- Consider a "G" or abstract AI-related icon

### After Creating:
The favicon will automatically appear in:
- Browser tabs
- Bookmarks
- Mobile home screen (when saved)
- Browser history

---

**Note:** Until you create the favicon files, browsers will show a default icon or no icon. This is normal and won't break functionality.
