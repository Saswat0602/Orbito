function sanitizeHtml(html: string, options?: any): string {
  if (!html) return html;
  return String(html).replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
}

sanitizeHtml.defaults = {
  allowedTags: ['p', 'div', 'h1', 'h2', 'h3', 'a', 'b', 'i', 'strong', 'em', 'span', 'img'],
  allowedAttributes: { a: ['href', 'name', 'target'], img: ['src', 'alt'] },
};

export = sanitizeHtml;

