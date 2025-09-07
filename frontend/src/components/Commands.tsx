import { useState } from "react";
import {
    getCommands,
    postCommands,
    type CommandsType,
} from "../api/shareClone/cmds";

const Commands = () => {
    const [cmd, setCmd] = useState<CommandsType>({
        id: "",
        name: "",
        commands: [{ name: "", description: "" }],
    });

    const [generatedId, setGeneratedId] = useState("");
    const [inputId, setInputId] = useState("");
    const [selected, setSelected] = useState(0);
    const [data, setData] = useState<CommandsType>();
    const [loading, setLoading] = useState(false);

    const handleGet = async () => {
        setLoading(true);
        try {
            const fetchedData = await getCommands(inputId);
            setData(fetchedData);
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
                cmd.name + "-" + Math.floor(Math.random() * 1000).toString();
            const postedData = await postCommands({ ...cmd, id });
            setGeneratedId(id);

            try {
                await navigator.clipboard.writeText(id);
                console.log("Copied!");
            } catch (err) {
                console.error("Copy failed:", err);
            }

            console.log("Success:", postedData);
        } catch (error) {
            console.error("Error:", error);
        }
    };

    const handleCmdChange = (
        index: number,
        field: "name" | "description",
        value: string,
    ) => {
        setCmd((prev) => {
            const newCmds = [...prev.commands];
            newCmds[index] = { ...newCmds[index], [field]: value };
            return { ...prev, commands: newCmds };
        });
    };

    const addCommand = () => {
        setCmd((prev) => ({
            ...prev,
            commands: [...prev.commands, { name: "", description: "" }],
        }));
    };

    return (
        <>
            <div className="flex space-x-2 pb-4 ">
                <button
                    onClick={() => setSelected(0)}
                    className={`transition-all duration-200 font-mono rounded-sm py-2 text-center ${
                        selected === 0
                            ? "px-8 bg-blue-400 text-white font-bold"
                            : "px-3 bg-white"
                    }`}
                >
                    Share
                </button>
                <button
                    onClick={() => setSelected(1)}
                    className={`transition-all duration-200 font-mono rounded-sm py-2 ${
                        selected === 1
                            ? "px-8 bg-pink-400 text-white font-bold"
                            : "px-3 bg-white"
                    }`}
                >
                    Clone
                </button>
            </div>

            <div className="">
                {selected === 0 && (
                    <div className="flex flex-col bg-zinc-100 h-96 w-11/12 p-4 space-y-2 overflow-y-auto">
                        <input
                            className="bg-zinc-300 p-2"
                            type="text"
                            placeholder="File name"
                            value={cmd.name}
                            onChange={(e) =>
                                setCmd((prev) => ({
                                    ...prev,
                                    name: e.target.value,
                                }))
                            }
                        />

                        {cmd.commands.map((c, index) => (
                            <div key={index} className="flex space-x-2">
                                <input
                                    className="flex-1 p-2"
                                    type="text"
                                    placeholder="Command name"
                                    value={c.name}
                                    onChange={(e) =>
                                        handleCmdChange(
                                            index,
                                            "name",
                                            e.target.value,
                                        )
                                    }
                                />
                                <input
                                    className="flex-1 p-2"
                                    type="text"
                                    placeholder="Description"
                                    value={c.description}
                                    onChange={(e) =>
                                        handleCmdChange(
                                            index,
                                            "description",
                                            e.target.value,
                                        )
                                    }
                                />
                            </div>
                        ))}

                        <button
                            onClick={addCommand}
                            className="transition-all duration-200 font-mono rounded-sm py-2 px-4 bg-green-400 text-white font-bold active:scale-105"
                        >
                            + Add Command
                        </button>

                        {generatedId && <h1>{generatedId}</h1>}

                        <button
                            onClick={handleSubmit}
                            className="transition-all duration-200 font-mono rounded-sm py-2 px-8 bg-blue-400 text-white font-bold text-center active:scale-105"
                        >
                            Get ID
                        </button>
                    </div>
                )}

                {selected === 1 && (
                    <div className="flex flex-col bg-zinc-100 w-11/12 p-4 space-y-2">
                        <input
                            className="bg-zinc-300 p-2"
                            type="text"
                            placeholder="Paste ID"
                            value={inputId}
                            onChange={(e) => setInputId(e.target.value)}
                        />
                        <button
                            onClick={handleGet}
                            className="transition-all duration-200 font-mono rounded-sm py-2 px-8 bg-blue-400 text-white font-bold text-center active:scale-105"
                        >
                            Get
                        </button>

                        {loading && <p>Loading...</p>}

                        {data && (
                            <>
                                <h1>{data.name}</h1>
                                {data.commands.map((c, index) => (
                                    <div key={index}>
                                        <p>{c.name}</p>
                                        <p>{c.description}</p>
                                    </div>
                                ))}
                            </>
                        )}
                    </div>
                )}
            </div>
        </>
    );
};

export default Commands;
