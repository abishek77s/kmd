export interface Command {
    name: string;
    description: string;
}

export interface CommandsType {
    id: string;
    name: string;
    commands: Command[];
}

const BASE_URL = "https://kmd-backend.onrender.com/commands";

export const getCommands = async (id: string): Promise<CommandsType> => {
    const res = await fetch(`${BASE_URL}/${id}`);
    if (!res.ok) throw new Error("Failed to fetch commands");
    return res.json();
};

export const postCommands = async (
    cmd: CommandsType,
): Promise<CommandsType> => {
    const res = await fetch(BASE_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(cmd),
    });
    if (!res.ok) throw new Error("Failed to post commands");
    return res.json();
};
