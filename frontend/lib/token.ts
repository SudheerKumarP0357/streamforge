export function getUserIdFromToken(token: string): string | null {
  // JWT is 3 base64 parts split by '.'
  // Decode the middle part (payload) to get claims
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = parts[1];
    const decodedStr = Buffer.from(payload, 'base64').toString('utf-8');
    const decoded = JSON.parse(decodedStr);
    
    return decoded.sub ?? null;
  } catch (err) {
    console.error('Failed to parse JWT token for user ID:', err);
    return null;
  }
}

export function getUserRoleFromToken(token: string): string | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = parts[1];
    const decodedStr = Buffer.from(payload, 'base64').toString('utf-8');
    const decoded = JSON.parse(decodedStr);
    
    return decoded.role ?? null;
  } catch (err) {
    console.error('Failed to parse JWT token for user role:', err);
    return null;
  }
}
