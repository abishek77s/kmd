import { useState } from "react";
import { type FileType, getFile, postFile } from "../api/shareClone/file";

const File = () => {
    const [file, setFile] = useState<FileType>({
        id: "",
        name: "",
        content: "",
    });

    const [inputId, setInputId] = useState("");
    const [selected, setSelected] = useState(0);

    const [generatedId, setGeneratedId] = useState("");
    const [data, setData] = useState<FileType>();
    const [loading, setLoading] = useState(false);

    const handleGet = async () => {
        setLoading(true);
        try {
            const fetchedFile = await getFile(inputId);
            setData(fetchedFile);
        } catch (err) {
            console.error("Error fetching:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleSubmit = async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        try {
            const id =
                file.name + "-" + Math.floor(Math.random() * 1000).toString();
            const postedFile = await postFile({ ...file, id });
            setGeneratedId(id);

            try {
                await navigator.clipboard.writeText(id);
                console.log("Copied!");
            } catch (err) {
                console.error("Copy failed:", err);
            }

            console.log("Success:", postedFile);
        } catch (error) {
            console.error("Error:", error);
        }
    };

    return (
        <>
            <div className="flex space-x-2 pb-4 ">
                <button
                    onClick={() => setSelected(0)}
                    className={`transition-all duration-200  font-mono rounded-sm py-2 text-center
              ${selected === 0 ? "px-8 bg-blue-400 text-white font-bold" : "px-3 bg-white"}`}
                >
                    Share
                </button>
                <button
                    onClick={() => setSelected(1)}
                    className={`transition-all duration-200 font-mono rounded-sm py-2
            ${selected === 1 ? "px-8 bg-pink-400 text-white font-bold" : "px-3 bg-white"}`}
                >
                    Clone
                </button>
            </div>
            <div className="">
                {selected == 0 && (
                    <div className="flex flex-col bg-zinc-100 h-96 w-11/12">
                        <input
                            className="bg-zinc-300"
                            type="text"
                            placeholder="File name"
                            onChange={(e) =>
                                setFile((prev) => ({
                                    ...prev,
                                    name: e.target.value,
                                }))
                            }
                        />
                        <textarea
                            className="h-full w-full p-2 border rounded"
                            onChange={(e) =>
                                setFile((prev) => ({
                                    ...prev,
                                    content: e.target.value,
                                }))
                            }
                            placeholder="Paste your code here..."
                            value={file.content}
                        />
                        {generatedId && <h1> {generatedId}</h1>}
                        <button
                            onClick={handleSubmit}
                            className={`transition-all duration-200 font-mono rounded-sm py-2 px-8
                        bg-blue-400 text-white font-bold text-center
                        active:scale-105`}
                        >
                            Get ID
                        </button>
                    </div>
                )}
                {selected == 1 && (
                    <div className="flex flex-col bg-zinc-100  w-11/12">
                        <input
                            className="bg-zinc-300"
                            onChange={(e) => setInputId(e.target.value)}
                            type="text"
                            placeholder="Paste ID"
                        />

                        <button
                            onClick={handleGet}
                            className={`transition-all duration-200 font-mono rounded-sm py-2 px-8
                      bg-blue-400 text-white font-bold text-center
                      active:scale-105`}
                        >
                            Get
                        </button>
                        {loading && <p> Loading...</p>}
                        <h1>{data?.name}</h1>
                        <pre className="whitespace-pre-wrap">
                            <code>{data?.content}</code>
                        </pre>
                    </div>
                )}
            </div>
        </>
    );
};

export default File;
