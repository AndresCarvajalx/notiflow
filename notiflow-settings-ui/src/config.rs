use std::fs;
use crate::model::Data;

const CONFIG_PATH: &str = "config.yml";

pub fn load() -> Result<Data, Box<dyn std::error::Error>> {
    let contenido = fs::read_to_string(CONFIG_PATH)?;
    let data: Data = serde_yml::from_str(&contenido)?;
    Ok(data)
}

pub fn save(data: &Data) -> Result<(), Box<dyn std::error::Error>> {
    let contenido = serde_yml::to_string(data)?;
    fs::write(CONFIG_PATH, contenido)?;
    Ok(())
}