export interface FileType {
    id: string;
    name: string;
    content: string;
}

const FILE_BASE_URL = "https://kmd-backend.onrender.com/file";

export const getFile = async (id: string): Promise<FileType> => {
    const res = await fetch(`${FILE_BASE_URL}/${id}`);
    if (!res.ok) throw new Error("Failed to fetch file");
    return res.json();
};

export const postFile = async (
    file: Omit<FileType, "id"> & { id?: string },
): Promise<FileType> => {
    const res = await fetch(FILE_BASE_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(file),
    });
    if (!res.ok) throw new Error("Failed to post file");
    return res.json();
};
