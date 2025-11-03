import React, { createContext, useContext, type ReactNode } from 'react';

interface LicenseContextType {
  licenseStatus: null;
  isLicenseValid: boolean;
  isLoading: boolean;
  error: string | null;
  checkLicenseStatus: () => Promise<void>;
}

const LicenseContext = createContext<LicenseContextType | undefined>(undefined);

interface LicenseProviderProps {
  children: ReactNode;
}

export const LicenseProvider: React.FC<LicenseProviderProps> = ({ children }) => {
  const checkLicenseStatus = async () => {
    // Empty stub - no license checking in open source version
  };

  const value: LicenseContextType = {
    licenseStatus: null,
    isLicenseValid: true, // Always valid in open source version
    isLoading: false,
    error: null,
    checkLicenseStatus,
  };

  return (
    <LicenseContext.Provider value={value}>
      {children}
    </LicenseContext.Provider>
  );
};

export const useLicense = (): LicenseContextType => {
  const context = useContext(LicenseContext);
  if (context === undefined) {
    throw new Error('useLicense must be used within a LicenseProvider');
  }
  return context;
}; 